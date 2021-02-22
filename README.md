# Yahtzee

This piece of software provides a basic backend service for playing yahtzee.

## API

**Every call** requires BASIC authentication. Users are not stored on the
backend; the `username` part of the header will be used as the player's name.

### List available features

```
GET /features
```

eg.
```
> GET /features
< 200 OK
< ["six-dice","yahtzee-bonus"]
```
### Create New Game

```
POST / < application/json [features...]
```

Available features are [here](#Features).

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

Available categories are [here](https://github.com/akarasz/yahtzee/blob/master/model.go#L22).

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
GET /{gameID}/hints
```

eg.
```
> GET /gcxog/hints
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

## Features
You can combine the features in any way you want, except for the Official and Yahtzee bonus. Official will overrule the Yahtzee bonus rules.

|Feature|Id|Description|
|-------|--|-----------|
|Official|`official`|Game with the [official yahtzee rules](https://en.wikipedia.org/wiki/Yahtzee#Rules).
|Yahtzee bonus|`yahtzee-bonus`|The base game with the yahtzee bonus and a basic joker rule: You are eligible for the bonus 100 point if you already filled the yahtzee with 50 points. You can also use your second yahtzee to fill the full house, small straight or large straight with 25-30-40 points respectively, without any restrictions.|
|Six dice|`six-dice`|Play the game with six dice. Points are calculated with the best 5 of them!|
|Ordered|`ordered`|Enforce top-down filling of the categories|
|Equilizer|`equilizer`|Everyone can score, except you? With the equilizer, when you score a zero in a category, all other players will have zero in the same category if they already filled that category. Use it wisely!|
|The Chance|`the-chance`|This is your chance to win! Score 5 points in the entire game, and you will get a bonus 495 at the end!|

## TODO

* store games in redis with an expiration
* use some kind of oauth (github? google?) instead of basic auth
