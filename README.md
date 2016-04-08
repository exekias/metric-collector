# metric-collector

This app showcases 3 workers processing metrics from a queue, metrics
have this format:

```
{
  "username": "fooser",  // string
  "count": 12412414,    // int64
  "metric": "kite_call" // string
}
```

To run backends and workers (in Docker) just run:

```
$ docker-compose up -d
Creating metriccollector_postgres_1
Creating metriccollector_redis_1
Creating metriccollector_rabbitmq_1
Creating metriccollector_distinctname_1
Creating metriccollector_accountname_1
Creating metriccollector_mongo_1
Creating metriccollector_hourlylog_1
```

You can check logs from all instances, or choose to view a worker:

```
$ docker-compose logs accountname
```

Then you can feed the system with random metrics running a test dispatcher:
```
$ go run dispatcher/main.go -debug
```

To kill & destroy the scenario just:
```
$ docker-compose kill
$ docker-compose rm
```
