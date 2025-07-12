## Schema

![](https://opendata.tweedekamer.nl/sites/default/files/styles/wide/public/images/OpenDataPortaal_InformatiemodelAlgemeen_3.png?itok=QTedYNs1)

## Info
etl can be ran using a crontab
```bash
0 2 * * * [path]/etl >> /var/log/etl.log 2>&1
```

make sure a postgres db exists and `DATABASE_URL` is set in env
