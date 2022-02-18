const {HelloRequest, RepeatHelloRequest,
			 HelloReply} = require('./echo_pb.js');
const {EchoClient} = require('./echo_grpc_web_pb.js');

var client = new EchoClient('https://' + window.location.hostname + ':8000',
															 null, null);

/* // simple unary call */
var request = new HelloRequest();
request.setName('Arion');

client.hello(request, {}, (err, response) => {
	if (err) {
		console.log(`Unexpected error for Hello: code = ${err.code}` +
								`, message = "${err.message}"`);
	} else {
		console.log(response.getMessage());
	}
});


// server streaming call
var streamRequest = new RepeatHelloRequest();
streamRequest.setName('Arion');
streamRequest.setCount(5);

var stream = client.repeatHello(streamRequest, {});
stream.on('data', (response) => {
	console.log(response.getMessage());
});
stream.on('error', (err) => {
	console.log(`Unexpected stream error: code = ${err.code}` +
							`, message = "${err.message}"`);
});
