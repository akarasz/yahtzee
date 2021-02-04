# Yahtzee

This piece of software provides a basic backend service for playing yahtzee.

## API

**Every call** requires BASIC authentication. Users are not stored on the
backend; the `username` part of the header will be used as the player's name.

### Create New Game

```
POST / < application/json [features...]
```

Available features are [here](https://github.com/nagymarci/yahtzee/blob/feature/six-dice/model.go#L100).

eg.
```
> POST / < ["six-dice"]
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
< {"Players": [
<   {
<     "User": "Alice",
<     "ScoreSheet": {}
<   }
< ]}
```

### Show a Game

```
GET /{gameID}
```

eg.
```
{
> GET /gcxog
< {
<   "Players":[
<     {
<       "User":"andris",
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
<   "RollCount":0,
<   "Features":[]
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
< {
<   "RollCount": 1,
<   "Dices": [
<     {
<       "Value": 1,
<       "Locked": true
<     },
<     {
<       "Value": 2,
<       "Locked": false
<     },
<     {
<       "Value": 3,
<       "Locked": false
<     },
<     {
<       "Value": 5,
<       "Locked": false
<     },
<     {
<       "Value": 1,
<       "Locked": true
<     }
<   ]
< }
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
< {"Dices": [
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
< ]}
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
< {
<   "Players":[
<     {
<       "User":"andris",
<       "ScoreSheet":{
<         "ones":3,
<         "small-straight":30,
<         "yahtzee":50
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
<   "Round":3,
<   "Current":0,
<   "RollCount":0
< }
```

### Score suggestions

```
GET /score?dices=[1-6],[1-6],[1-6],[1-6],[1-6]
```

eg.
```
> POST /score?dices=2,3,1,3,2
< 200 OK
< {
<   "ones": 1,
<   "twos": 4,
<   "threes": 6,
<   "fours": 0,
<   "fives": 0,
<   "sixes": 0,
<   "three-of-a-kind": 0,
<   "four-of-a-kind": 0,
<   "full-house": 0,
<   "small-straight": 0,
<   "large-straight": 0,
<   "yahtzee": 0,
<   "chance": 11,
< }
```

## TODO

* store games in redis with an expiration
* use some kind of oauth (github? google?) instead of basic auth
