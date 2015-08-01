#GO Online Judge

###GET DEPENDENCIES

`go get github.com/ggaaooppeenngg/OJ`

###RUN PROJECT

Modify **conf/app.conf** and **conf/misc.conf** to adapt to your enviroment.

This includes two parts, one is web server, another is judger,
judger scans database to process unhandled submits.

To run the server:

```

revel run github.com/ggaaooppeenngg/OJ 

```

To run the judger:

```
go build -o judge/judge judge/judge.go

./judge/judge

```

####SET UP SANDBOX FOR CODE TESIING

Sandbox is an independent package.

```

go get github.com/ggaaooppeenngg/sandbox
go install github.com/ggaaooppeenngg/sandbox

```

For more details, see [here](http://github.com/ggaaooppeenngg/sandbox)

