# dodualm

API and tooling for the Dofus Almanax.

## Setup

Create the database with the following command:
```bash
dodualm migrate up
```

Now you can start the server with:
```bash
dodualm
```

## How it works

Future Almanax data is volatile while past data is static. Almanax data is only updated with client updates to the quest data. While there is a common pattern, each update can override it.
Dodualm is an API that is completely generated from the dofusdude data repository. It is a twin to doduapi with persistent data on top.
On every update, dodualm is notified by the dofusdude pipeline and it fetches the newest updates to the future data and updates the database.

## License
[MIT](https://choosealicense.com/licenses/mit/)
