1. Put some xml files in the ./seeds directory
2. then...

```sh
npm install
docker-compose up
curl localhost:5555/api/cache/seed
open localhost:5555/ui
```