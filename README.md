# NHL Party Score

An app to watch the NHL API and take action (ring a bell, turn a light on) when a goal is scored.

Based on [https://gitlab.com/dword4/nhlapi](https://gitlab.com/dword4/nhlapi)

## Score API

```
curl -X 'GET' -kvs   'https://api-web.nhle.com/v1/score/2025-04-15'   -H 'accept: application/json' |jq '.' > score-api.json
```

## List teams

use curl to call the API

```
curl -X 'GET' -kvs   'https://api-web.nhle.com/v1/score/2025-04-16'   -H 'accept: application/json' |jq -r '.games.[]| [.startTimeUTC, .homeTeam.id, .homeTeam.name.default,  .awayTeam.id, .awayTeam.name.default ] |@tsv '
```
