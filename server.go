package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	mapping "github.com/dofusdude/dodumap"
	"github.com/google/go-github/v67/github"
	"github.com/meilisearch/meilisearch-go"
)

var (
	DataRepoOwner         = "dofusdude"
	DataRepoName          = "dofus3-main"
	MappedAlmanaxFileName = "MAPPED_ALMANAX.json"
	Languages             = []string{"en", "fr", "de", "es", "pt"}
)

func loadAlmanaxData(version string) ([]mapping.MappedMultilangNPCAlmanax, error) {
	client := github.NewClient(nil)

	repRel, _, err := client.Repositories.GetReleaseByTag(context.Background(), DataRepoOwner, DataRepoName, version)
	if err != nil {
		return nil, err
	}

	// get the mapped almanax data
	var assetId int64
	assetId = -1
	for _, asset := range repRel.Assets {
		if asset.GetName() == MappedAlmanaxFileName {
			assetId = asset.GetID()
			break
		}
	}

	if assetId == -1 {
		return nil, fmt.Errorf("could not find asset with name %s", MappedAlmanaxFileName)
	}

	log.Info("downloading asset", "assetId", assetId)
	httpClient := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Automatically follow all redirects
			return nil
		},
	}
	asset, redirectUrl, err := client.Repositories.DownloadReleaseAsset(context.Background(), DataRepoOwner, DataRepoName, assetId, httpClient)
	if err != nil {
		return nil, err
	}

	if asset == nil {
		return nil, fmt.Errorf("asset is nil, redirect url: %s", redirectUrl)
	}

	defer asset.Close()

	var almData []mapping.MappedMultilangNPCAlmanax
	dec := json.NewDecoder(asset)
	err = dec.Decode(&almData)
	if err != nil {
		return nil, err
	}

	return almData, nil
}

/*
*
per default the current day almanax in the requested language
timezone paris

query params:
- range[start_date] - start date in format yyyy-mm-dd, default today
- range[end_date] - end date in format yyyy-mm-dd, default today (inclusive)
- timezone - timezone to use, default Europe/Paris
- filter[bonus.type_name] - filter bonuses by type name, off by default
- filter[bonus.id], english name for the bonus, off by default
- query[bonus.name] - search for bonuses by localized name and directly return the bonuses sorted by date
*/
func RetrieveAlmanax(w http.ResponseWriter, r *http.Request) {
	// TODO
}

// TODO use own db instead of alm/bonuses endpoint
func UpdateAlmanaxBonusIndex(init bool) int {
	client := meilisearch.New(MeiliHost, meilisearch.WithAPIKey(MeiliKey))
	defer client.Close()

	added := 0

	for _, lang := range Languages {
		if lang == "pt" {
			continue // no portuguese almanax bonuses
		}
		url := fmt.Sprintf("https://api.dofusdu.de/dofus2/meta/%s/almanax/bonuses", lang)
		resp, err := http.Get(url)
		if err != nil {
			log.Warn(err, "lang", lang)
			return added
		}

		var bonuses []AlmanaxBonusListing
		err = json.NewDecoder(resp.Body).Decode(&bonuses)
		if err != nil {
			log.Error(err, "lang", lang)
			return added
		}

		var bonusesMeili []AlmanaxBonusListingMeili
		var counter int = 0
		for i := range bonuses {
			bonusesMeili = append(bonusesMeili, AlmanaxBonusListingMeili{
				Id:   strconv.Itoa(counter),
				Slug: bonuses[i].Id,
				Name: bonuses[i].Name,
			})
			counter++
		}

		indexName := fmt.Sprintf("alm-bonuses-%s", lang)
		_, err = client.GetIndex(indexName)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				log.Info("alm bonuses index does not exist yet, creating now", "index", indexName)
				almTaskInfo, err := client.CreateIndex(&meilisearch.IndexConfig{
					Uid:        indexName,
					PrimaryKey: "id",
				})

				if err != nil {
					log.Error("Error while creating alm bonus index in meili", "err", err)
					return added
				}

				task, err := client.WaitForTask(almTaskInfo.TaskUID, 500*time.Millisecond)
				if err != nil {
					log.Error("Error while waiting alm bonus index creation at meili", "err", err)
					return added
				}

				if task.Status == "failed" && !strings.Contains(task.Error.Message, "already exists") {
					log.Error("alm bonuses creation failed.", "err", task.Error)
					return added
				}

			} else {
				log.Error("Error while getting alm bonus index in meili", "err", err)
				return added
			}
		}

		almBonusIndex := client.Index(indexName)

		if init { // clean index, add all
			cleanTask, err := almBonusIndex.DeleteAllDocuments()
			if err != nil {
				log.Error("Error while cleaning alm bonuses in meili.", "err", err)
				return added
			}

			task, err := client.WaitForTask(cleanTask.TaskUID, 100*time.Millisecond)
			if err != nil {
				log.Error("Error while waiting for meili to clean alm bonuses.", "err", err)
				return added
			}

			if task.Status == "failed" {
				log.Error("clean alm bonuses task failed.", "err", task.Error)
				return added
			}

			var documentsAddTask *meilisearch.TaskInfo
			if documentsAddTask, err = almBonusIndex.AddDocuments(bonusesMeili); err != nil {
				log.Error("Error while adding alm bonuses to meili.", "err", err)
				return added
			}

			task, err = client.WaitForTask(documentsAddTask.TaskUID, 500*time.Millisecond)
			if err != nil {
				log.Error("Error while waiting for meili to add alm bonuses.", "err", err)
				return added
			}

			if task.Status == "failed" {
				log.Error("alm bonuses add docs task failed.", "err", task.Error)
				return added
			}

			added += len(bonuses)
		} else { // search the item exact matches before adding it
			for _, bonus := range bonusesMeili {
				request := &meilisearch.SearchRequest{
					Limit: 1,
				}

				var searchResp *meilisearch.SearchResponse
				if searchResp, err = almBonusIndex.Search(bonus.Name, request); err != nil {
					log.Error("SearchAlmanaxBonuses: index not found: ", "err", err)
					return added
				}

				foundIdentical := false
				if len(searchResp.Hits) > 0 {
					var item = searchResp.Hits[0].(map[string]interface{})
					if item["name"] == bonus.Name {
						foundIdentical = true
					}
				}

				if !foundIdentical { // add only if not found
					log.Info("adding", "bonus", bonus.Name, "bonus", bonus, "lang", lang, "hits", searchResp.Hits)
					documentsAddTask, err := almBonusIndex.AddDocuments([]AlmanaxBonusListingMeili{bonus})
					if err != nil {
						log.Error("Error while adding alm bonuses to meili.", "err", err)
						return added
					}

					task, err := client.WaitForTask(documentsAddTask.TaskUID, 500*time.Millisecond)
					if err != nil {
						log.Error("Error while waiting for meili to add alm bonuses.", "err", err)
						return added
					}

					if task.Status == "failed" {
						log.Error("alm bonuses adding failed.", "err", task.Error)
						return added
					}

					added += 1
				}
			}
		}
	}

	return added
}

type UpdateAlmanaxRequest struct {
	// Load the mapped_almanax on startup, update with doduda API request, reload date => npc pairs from alm-dates repo.
	// alm-dates runs short.sh etc and manages files for each year. scripts update the <year>.json if something changes.
	// UpdateAlmanaxRequest then reads the files and updates the mapped_almanax.
}

func UpdateAlmanax(w http.ResponseWriter, r *http.Request) {
	// TODO
}

func ListBonuses(w http.ResponseWriter, r *http.Request) {
	// TODO
}

func getLimitInBoundary(limitStr string) (int64, error) {
	if limitStr == "" {
		limitStr = "8"
	}
	var limit int
	var err error
	if limit, err = strconv.Atoi(limitStr); err != nil {
		return 0, fmt.Errorf("invalid limit value")
	}
	if limit > 100 {
		return 0, fmt.Errorf("limit value is too high")
	}

	return int64(limit), nil
}

func SetJsonHeader(w *http.ResponseWriter) {
	(*w).Header().Set("Content-Type", "application/json")
}

func WriteCacheHeader(w *http.ResponseWriter) {
	SetJsonHeader(w)
	//(*w).Header().Set("Cache-Control", "max-age:300, public")
	//(*w).Header().Set("Last-Modified", LastUpdate.Format(http.TimeFormat))
	//(*w).Header().Set("Expires", time.Now().Add(time.Minute*5).Format(http.TimeFormat))
}

func SearchBonuses(w http.ResponseWriter, r *http.Request) {
	client := meilisearch.New(MeiliHost, meilisearch.WithAPIKey(MeiliKey))
	defer client.Close()

	query := r.URL.Query().Get("query")
	if query == "" {
		writeInvalidQueryResponse(w, "Query parameter is required.")
		return
	}

	lang := r.Context().Value("lang").(string)

	if lang == "pt" {
		writeInvalidQueryResponse(w, "Portuguese language is not translated for Almanax Bonuses.")
		return
	}

	var searchLimit int64
	var err error
	if searchLimit, err = getLimitInBoundary(r.URL.Query().Get("limit")); err != nil {
		writeInvalidQueryResponse(w, "Invalid limit value: "+err.Error())
		return
	}

	index := client.Index(fmt.Sprintf("alm-bonuses-%s", lang))

	request := &meilisearch.SearchRequest{
		Limit: searchLimit,
	}

	var searchResp *meilisearch.SearchResponse
	if searchResp, err = index.Search(query, request); err != nil {
		writeServerErrorResponse(w, "Could not search: "+err.Error())
		return
	}

	//requestsTotal.Inc()
	//requestsSearchTotal.Inc()

	if searchResp.EstimatedTotalHits == 0 {
		writeNotFoundResponse(w, "No results found.")
		return
	}

	var results []AlmanaxBonusListing
	for _, hit := range searchResp.Hits {
		almBonusJson := hit.(map[string]interface{})
		almBonus := AlmanaxBonusListing{
			Id:   almBonusJson["slug"].(string),
			Name: almBonusJson["name"].(string),
		}
		results = append(results, almBonus)
	}

	WriteCacheHeader(&w)
	err = json.NewEncoder(w).Encode(results)
	if err != nil {
		writeServerErrorResponse(w, "Could not encode JSON: "+err.Error())
		return
	}
}
