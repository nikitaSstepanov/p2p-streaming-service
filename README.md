<h1>Hi there 👋</h1>
<h1>It is p2p streaming service</h1>

### Description

It is streaming service based on bittorrent. We take the data from the torrent.

This service has two parts: website and backend. 

It allows users to watch movies available on the service, create playlists. Administrators can add new movies and edit old ones.

Developers: 

   1) Backend: <a href="https://github.com/Reprr">Lev</a>, <a href="https://github.com/nikitaSstepanov">Nikita</a>
   2) Frontend: <a href="https://github.com">Ural</a>

P.S. We don`t use piratical resources. We respect copyrights and urge you not to use the service with stolen movies and other resources.

### Backend

Our backend application and bittorent client are written in golang. 

When running the application in docker, requests are proxied using nginx.

Launch:
   
   All paths specified in this section are relative to the "backend" folder
   
   1) Setup configuration of the project (config/config.yml) or use default values.
      
   2) If you don`t use docker, raise the postgres and redis databases in advance (do not forget that app will try to connect to the db with the environment specified in config/config.yml).
      
   3) Set up the environment (create ".env" file) according to the example in the ".env.example" file or rename ".env.example" to ".env" to use default values.
   
   4) Navigate to the "backend" folder in the terminal (relative to the root of the project) and enter the command:

      ```docker-compose up```

   5) If you don`t use docker, enter:

      ```go run ./cmd/p2p-streaming-service/main.go```

Usage:

Usage will be later...

### Frontend

Frontend will be later...

### Stack

<p>
    <a href="https://go.dev">
        <img width=70 alt="GO" src="https://logodix.com/logo/2142682.png"/>
    </a>
    <a href="https://www.docker.com">
        <img width="50" alt="Docker" src="https://logodix.com/logo/826596.png"/>
    </a>
    <a href="https://www.postgresql.org">
        <img width="50" alt="PG" src="https://logodix.com/logo/2106569.png">
    </a>
    <a href="https://nginx.org">
        <img width="70" alt="Nginx" src="https://logodix.com/logo/1638878.png"/>
    </a>
    <a href="https://redis.io">
        <img width="70" alt="Redis" src="https://logodix.com/logo/631151.png"/>
    </a>
</p>


