# plexarr

Simple CLI tool to fix Plex library matches according to Sonarr/Radarr

## Sample Configuration

```yml
plex:
  url: https://plex.domain.com
  token: your-plex-token
  database: /opt/plex/Library/Application Support/Plex Media Server/Plug-in Support/Databases/com.plexapp.plugins.library.db

pvr:
  radarr:
    - name: radarr
      url: https://radarr.domain.com
      api_key: your-radarr-token
      rewrite:
        from: /mnt/unionfs/Media/*
        to: /data/$1
        
  sonarr:
    - name: sonarr
      url: https://sonarr.domain.com
      api_key: your-sonarr-token
      rewrite:
        from: /mnt/unionfs/Media/*
        to: /data/$1
```

## Sample Commands

`plexarr --pvr sonarr --library TV`

`plexarr --pvr radarr --library Movies`

`plexarr --pvr radarr --library Movies-Action --library Movies-Comedy`

