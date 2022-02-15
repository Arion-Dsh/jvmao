///
//  Generated code. Do not modify.
//  source: echo.proto
//
// @dart = 2.12
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:async' as $async;

import 'dart:core' as $core;

import 'package:grpc/service_api.dart' as $grpc;
import 'echo.pb.dart' as $0;
export 'echo.pb.dart';

class EchoClient extends $grpc.Client {
  static final _$hello = $grpc.ClientMethod<$0.HelloRequest, $0.HelloReply>(
      '/proto.Echo/Hello',
      ($0.HelloRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $0.HelloReply.fromBuffer(value));

  EchoClient($grpc.ClientChannel channel,
      {$grpc.CallOptions? options,
      $core.Iterable<$grpc.ClientInterceptor>? interceptors})
      : super(channel, options: options, interceptors: interceptors);

  $grpc.ResponseFuture<$0.HelloReply> hello($0.HelloRequest request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$hello, request, options: options);
  }
}

abstract class EchoServiceBase extends $grpc.Service {
  $core.String get $name => 'proto.Echo';

  EchoServiceBase() {
    $addMethod($grpc.ServiceMethod<$0.HelloRequest, $0.HelloReply>(
        'Hello',
        hello_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.HelloRequest.fromBuffer(value),
        ($0.HelloReply value) => value.writeToBuffer()));
  }

  $async.Future<$0.HelloReply> hello_Pre(
      $grpc.ServiceCall call, $async.Future<$0.HelloRequest> request) async {
    return hello(call, await request);
  }

  $async.Future<$0.HelloReply> hello(
      $grpc.ServiceCall call, $0.HelloRequest request);
}
