## JvMao 

jvmao(橘猫) a kind of cat in China. 

grace powerful, but easy to get fat. 

look like:

<img src="jvmao.webp" width="300">

## useage

    jm := jvmao.New()
	jm.Use(middleware.Logger())

    jm.GET("home", "/", func(c jvmao.Context)error{
       return c.String(http.SatusOK, "home page")
    })

    jm.Start(":8000")


## gRPC-web
 as we know, gRPC-web clients connect to gRPC services via a special proxy. such as Envoy.
 jvmao improve a mini proxy for gRPC-web

    serv := grpc.NewServer(opts...)

    jm := jvmao.New()
    jm.RegisterGrpcServer(serv)


find the full code in examples.






