# googleroutes

A example of how to use the [Google maps directions
API](https://developers.google.com/maps/documentation/directions/start).  This
prints out the estimated travel time between two spots every 10m for the next
day.

You need an [API
Key](https://developers.google.com/maps/documentation/directions/start#api-key).

[Full API
docs](https://developers.google.com/maps/documentation/directions/get-directions)

```
% go run googleroutes.go --source 53.339710,-6.237448  -destination 51.516116,-0.127209 --api_key <KEY> --travel_model=DRIVE
```

