package main

// Windows doesn't ship with a zoneinfo database. Without this LoadLocation fails.
import _ "time/tzdata"
