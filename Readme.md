Track yo'self
=============

This set of programs tracks your i3 window focus and stores it in a mongodb
database running on the local host.


Usage
-----

`tracking` is meant to be running at all times. It queries i3 through `i3-msg`
every 10 seconds and stores the focused window's information in the database.
You can simply add it with `exec /path/to/tracking` in your i3 configuration
file.

`gen_graph` creates a pie chart, `gen_chart` creates a timeline showing one day
per column with colored stripes representing a focused window. The usage for
both of these programs is:

`gen_<chart|graph> <amount> [modifier]`

`modifier` specifies the scale `amount` is multiplied with. Possible values
are: `S` or `s` for seconds, `M` or `m` for minutes, `H` or `h` for hours
and `D` or `d` for days, which is also the default modifier. The amount is
subtracted from the current day and data is taken starting from that day.

For example:

`gen_chart 10 d`

will create a timeline chart for the last 10 days.

Dependencies
------------

You will need these packages:

- go.mongodb.org/mongo-driver/bson
- go.mongodb.org/mongo-driver/mongo
- go.mongodb.org/mongo-driver/mongo/options
- github.com/fogleman/gg
- github.com/wcharczuk/go-chart

Get them by running:

```shell
go get go.mongodb.org/mongo-driver/bson
go get go.mongodb.org/mongo-driver/mongo
go get github.com/fogleman/gg
go get github.com/wcharczuk/go-chart
```

or just use the supplied `go.mod` file:

```shell
go mod download
```

Misc.
-----

I wrote this while still learning go. There is probably a lot non-ideomatic and
inefficient code. However, it seems to mostly work and maybe someone else is
also interested in tracking their computer usage. I am not familiar with how
other window managers can be queried, and I am not interested in porting it
myself. Pull requests welcome!

Todo/Wishlist
-------------

- cleaner code
- better selectable date range
- even fancier graphs
- ...
