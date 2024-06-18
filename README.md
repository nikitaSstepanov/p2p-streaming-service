<h1>Hi there 👋</h1>
<h1>It is p2p streaming service</h1>

### Description

It is streaming service based on bittorrent. We take the data from the torrent.

It allows users to watch movies available on the service, create playlists and do a lot more. Administrators can add new movies and edit old ones.

Developers: 

   1) Backend: <a href="https://github.com/nikitaSstepanov">Nikita</a>, <a href="https://github.com/Reprr">Lev</a>

P.S. We don`t use piratical resources. We respect copyrights and urge you not to use the service with stolen movies and other resources.

### Backend

Our backend application and bittorent client are written in golang. 

When running the application in docker, requests are proxied using nginx.

Launch:
   
   All paths specified in this section are relative to the "backend" folder
   
   1) Set up configuration of the project (configs/config.yml) or use default values.

   2) Set up configuration of docker (docker-compose.yml) if you use it.
      
   3) If you don`t use docker, raise the postgres and redis databases in advance (do not forget that app will try to connect to the dbs with the environment specified in configs/config.yml).
      
   4) Set up the environment (create ".env" file) according to the example in the ".env.example" file or rename ".env.example" to ".env" to use default values.
   
   5) Navigate to the "backend" folder in the terminal (relative to the root of the project) and enter the command:

      ```docker-compose up```
    
      If you don`t use docker, enter:

      ```go run ./cmd/p2p-streaming-service/main.go```

Usage:

To learn usage of backend you can go to OpenAPI (backend/docs/openapi.yml).
