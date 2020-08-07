# Yahtzee

This piece of software provides a basic backend service for playing yahtzee.

## API

**Every call** requires BASIC authentication. Users are not stored on the
backend; the `username` part of the header will be used as the player's name.

### Create New Game

```
POST /
```

eg.
```
> POST /
< 201 Created
< Location: /{gameID}
```

### Join an Existing Game

```
POST /{gameID}/join
```

eg.
```
> POST /gcxog/join
< 201 Created
```

### Show a Game

```
GET /{gameID}
```

eg.
```
{
> GET /gcxog
<   "Players":[
<     {
<       "Name":"andris",
<       "ScoreSheet":{
<         "ones":3,
<         "small-straight":30
<       }
<     }
<   ],
<   "Dices":[
<     {
<       "Value":6,
<       "Locked":false
<     },
<     {
<       "Value":1,
<       "Locked":false
<     },
<     {
<       "Value":3,
<       "Locked":false
<     },
<     {
<       "Value":5,
<       "Locked":false
<     },
<     {
<       "Value":4,
<       "Locked":false
<     }
<   ],
<   "Round":2,
<   "Current":0,
<   "RollCount":0
< }
```

### Roll the dices

```
POST /{gameID}/roll
```

eg.
```
> POST /gcxog/roll
< 200 OK
<
< [
<   {
<     "Value": 1,
<     "Locked": true
<   },
<   {
<     "Value": 2,
<     "Locked": false
<   },
<   {
<     "Value": 3,
<     "Locked": false
<   },
<   {
<     "Value": 5,
<     "Locked": false
<   },
<   {
<     "Value": 1,
<     "Locked": true
<   }
< ]
```

### Toggle Lock on a Dice

```
POST /{gameID}/lock/{diceIndex}
```

eg.
```
> POST /gcxog/lock/3
< 200 OK
<
< [
<   {
<     "Value": 1,
<     "Locked": true
<   },
<   {
<     "Value": 2,
<     "Locked": false
<   },
<   {
<     "Value": 3,
<     "Locked": false
<   },
<   {
<     "Value": 5,
<     "Locked": true
<   },
<   {
<     "Value": 1,
<     "Locked": true
<   }
< ]
```

### Score

```
POST /{gameID}/score < text/plain `category`
```

Available categories are [here](https://github.com/akarasz/yahtzee/blob/master/pkg/game/game.go#L22).

eg.
```
> POST /gcxog/score < `yahtzee`
< 200 OK
```

## TODO

* handler tests
* better logging
* heroku deployment
* integrate logging with logdna
* websocket for announcing real time state changes
* store should return concrete object and not pointer (pointer only works for
  in-memory, not for redis)
* store games in redis with an expiration
* use some kind of oauth (github? google?) instead of basic auth
