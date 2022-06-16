# JSON data files

** WARNING: this will likely change in the next release **

For 'historical reasons' (see [Notes](#notes)), the backing data for the `uhppoted-httpd` server is stored as
a set of JSON files:

- `interfaces.json`
- `controllers.json`
- `doors.json`
- `cards.json`
- `groups.json`
- `users.json`
- `events.json`
- `logs.json`
- `history.json`

in the _var_ folder:

- /var/uhppoted/httpd/system (Linux)
- /usr/local/var/com.github.uhppoted/httpd/system (MacOS)
- \Program Data\uhppoted\httpd\system (Windows)

## Data structure

### `interfaces.json`
```
{
  "interfaces": [
    {
      "OID": "0.1.1",
      "name": "LANICA",
      "bind-address": "192.168.1.100",
      "broadcast-address": "192.168.1.255",
      "listen-address": "192.168.1.100:60001",
      "created": "2021-11-18 20:33:38 UTC",
      "modified": "2022-05-13 17:38:10 UTC"
    }
  ]
}
```

### `controllers.json`
```
{
  "controllers": [
    {
      "OID": "0.2.2",
      "name": "Beta",
      "device-id": 303986753,
      "address": "192.168.1.100",
      "doors": {
        "1": "0.3.1",
        "2": "0.3.2",
        "3": "0.3.3",
        "4": "0.3.4"
      },
      "timezone": "UTC",
      "created": "2022-04-02 18:32:51 UTC",
      "modified": "2022-05-19 16:18:02 UTC"
    },
    ...
  ]
}
```

### `doors.json`
```
{
  "doors": [
    {
      "OID": "0.3.7",
      "name": "Ravenclaw",
      "delay": 7,
      "mode": "normally open",
      "created": "2021-08-17 20:12:05 UTC",
      "modified": "2022-06-07 18:12:59 UTC"
    },
    ...
  ]
}
```

### `cards.json`
```
{
  "cards": [
    {
      "OID": "0.4.1",
      "name": "Albus Dumbledore",
      "card": 8000001,
      "from": "2022-02-01",
      "to": "2022-12-31",
      "groups": [
        "0.5.1",
        "0.5.8",
        "0.5.2"
      ],
      "created": "2022-04-05 18:41:28 UTC",
      "modified": "2022-05-31 18:08:34 UTC"
    },
    ...
  ]
}
```

### `groups.json`
```
{
  "groups": [
    {
      "OID": "0.5.7",
      "name": "Slytherin",
      "doors": [
        "0.3.8"
      ],
      "created": "2021-09-06 19:01:27 UTC",
      "modified": "2022-05-20 15:54:52 UTC"
    },
    ...
  ]
}
```

### `users.json`
```
{
  "users": [
    {
      "OID": "0.8.1",
      "name": "David Duque",
      "uid": "admin",
      "role": "admin",
      "salt": "4a3d95eec4b94c236824f75ef8305c90",
      "password": "91b228c1595768ff35f8bce555ac089e8e64348e15110fc8ef3653fd4cf7ff67",
      "created": "2022-02-10 18:41:48 UTC",
      "modified": ""
    },
    ...
  ]
}
```

### `events.json`
```
{
  "events": [
    {
      "OID": "0.6.100",
      "device-id": 405419896,
      "index": 68,
      "timestamp": "2021-08-10 10:28:28 PDT",
      "event-type": 2,
      "door": 4,
      "direction": 1,
      "card": 8165536,
      "granted": true,
      "reason": 1,
      "device-name": "AlphaQ"
    },
    ...
  ]
}
```

### `logs.json`
```
{
  "logs": [
    {
      "timestamp": "2022-05-24T11:47:07.186876-07:00",
      "UID": "admin",
      "OID": "0.7.100677",
      "item": "card",
      "id": "6000001",
      "name": "Harry PotterX",
      "field": "name",
      "details": "Updated name from 'Harry PotterX' to 'Harry Potter'"
    },
    ...
  ]
}
```

### `history.json`
```
{
  "history": [
    {
      "timestamp": "2022-06-10T10:52:46.006024-07:00",
      "item": "door",
      "id": "405419896/4",
      "field": "name",
      "before": "SlytherinX",
      "after": "Slytherin"
    },
    ...
  ]
}
```

## Notes

The design decisions that led to the system data being stored in a set of JSON files are more than likely
going to be revisited in the near future - but FWIW, here are the 'historical reasons':

1. The primary reason is that one of the original design goals was to keep external dependencies to an 
   absolute minimum. At the time of writing there are still only two real dependencies:

   - the [JWT library](https://github.com/cristalhq/jwt/v3)
   - the [grule-rule-engine](https://github.com/hyperjumptech/grule-rule-engine)

   both of which could be replaced with alternatives without any major structural impact. A DB (with or
   without an ORM) would have inevitably become a core dependency with subtle and limiting effects on 
   architectural and implementation choices.

2. The amount of system data was always envisaged as being relatively small so an in-memory database was 
   always practical.

3. Choosing e.g. a SQL/NoSQL database at the outset seemed likely to limit some architectural choices as 
   well as making some detailed implementation unnecessarily complicated and arcane. In particular, one 
   of the design goals was to be able to easily access and update individual items from the user interface
   without requiring detailed knowledge of the internal structure of the data.

4. Starting out, the natural structure of the data was not clear - it seemed to be an unholy mix of tree, table
   and ad hoc and from experience, shoe-horning it into a set of regular DB tables was potentially problematic.





