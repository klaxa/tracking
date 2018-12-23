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
and `D` or `d` for days, which is also the default modifier.

