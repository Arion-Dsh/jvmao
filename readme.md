## JvMao 

jvmao(橘猫) a kind of cat in China. 

grace powerful, but easy to get fat. 

look like:

<img src="jvmao.webp" width="300">

## useage

    jm := jvmao.New()
	jm.Use(middleware.Logger())

    jm.GET("home", "/", func(c *jvmao.Context)error{
       return c.String(http.SatusOK, "home page")
    })

    jm.Start(":8000")


## notice
unstable now.

