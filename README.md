# NHL Party Score

An app to watch the NHL API and take action (ring a bell, turn a light on) when a goal is scored.

This app was mostly fully generated using Claude AI with a single prompt. It was a "night project".
The goal is to get the app working on an ESP32 or Arduino, with a connected (and 3D printed) light. But right now you have to run it in your terminal.

This App is using the public NHL API, that is described at [https://gitlab.com/dword4/nhlapi](https://gitlab.com/dword4/nhlapi)

## Setup

Simply build as a regular Go app:

```bash
go get -u ./...
go build
```

Then rename `config-sample.yaml` to `config.yaml` and edit it to your needs.

Finaly, run the app:

```bash
./nhlPartyScore
```

### Hue Setup

- go to [https://discovery.meethue.com/](https://discovery.meethue.com/) to get the IP address of your Hue Bridge
- follow [https://developers.meethue.com/develop/get-started-2/](https://developers.meethue.com/develop/get-started-2/) to get the username of your Hue Bridge


## Score API

```
curl -X 'GET' -kvs   'https://api-web.nhle.com/v1/score/2025-04-15'   -H 'accept: application/json' |jq '.' > score-api.json
```

## List teams

use curl to call the API

```
curl -X 'GET' -kvs   'https://api-web.nhle.com/v1/score/2025-04-16'   -H 'accept: application/json' |jq -r '.games.[]| [.startTimeUTC, .homeTeam.id, .homeTeam.name.default,  .awayTeam.id, .awayTeam.name.default ] |@tsv '
```
