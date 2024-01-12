import 'dart:async';
import 'dart:convert';
import 'dart:io';

// These should be copied into place when we run 'make generate' for testing.
import 'sample_service.gen.client.dart';

main(List<String> args) async {
  if (args.length < 2) {
    print('Usage:');
    print('  dart run_client.dart {ADDRESS} {TEST_CASE}');
    print('');
    print('Example:');
    print('  dart run_client.dart localhost:9000 DownloadResumable');
    print('  dart run_client.dart localhost:9001 Authorization');
    exit(1);
  }
  await runTestCase(args[0], args[1]);
  exit(0);
}

runTestCase(String hostPort, String name) async {
  var baseURI = 'http://' + hostPort;

  switch (name) {
    case 'NotConnected':
      var client = new SampleServiceClient(baseURI);
      return output(client.Defaults(SampleRequest(text: 'Dude')));

    case 'Defaults':
      var client = new SampleServiceClient(baseURI);
      return output(client.Defaults(SampleRequest(text: 'Dude')));

    case 'ComplexValues':
      var client = new SampleServiceClient(baseURI);
      return output(client.ComplexValues(SampleComplexRequest(
        inFlag: true,
        inFloat: 3.14,
        inTime: '2022-12-05T17:47:12+00:00',
        inTimePtr: '2020-11-06T17:47:12+00:00',
        inUser: SampleUser(
          id: '123',
          name: 'Dude',
          age: 47,
          attention: 1000000,
          attentionString: '4m2s',
          digits: '555-1234',
          marshalToString: 'home@string.com,work@string.com',
          marshalToObject: {
            'H': 'home@object.com',
            'W': 'work@object.com',
          },
        ),
      )));

    case 'ComplexValuesPath':
      var client = new SampleServiceClient(baseURI);
      return output(client.ComplexValuesPath(SampleComplexRequest(
        inFlag: true,
        inFloat: 3.14,
        inTime: '2022-12-05T17:47:12+00:00',
        inTimePtr: '2020-11-06T17:47:12+00:00',
        inUser: SampleUser(
          id: '123',
          name: 'Dude',
          age: 47,
          attention: 1000000,
          attentionString: '4m2s',
          digits: '555-1234',
          marshalToString: 'home@string.com,work@string.com',
          marshalToObject: {
            'Home': "home@object.com",
            'Work': "work@object.com",
          },
        ),
      )));

    case 'Fail4XX':
      var client = new SampleServiceClient(baseURI);
      return output(client.Fail4XX(SampleRequest()));

    case 'Fail5XX':
      var client = new SampleServiceClient(baseURI);
      return output(client.Fail5XX(SampleRequest()));

    case 'CustomRoute':
      var client = new SampleServiceClient(baseURI);
      return output(client.CustomRoute(SampleRequest(id: '123', text: 'Abide')));

    case 'CustomRouteBody':
      var client = new SampleServiceClient(baseURI);
      return output(client.CustomRouteBody(SampleRequest(id: '123', text: 'Abide')));

    case 'CustomRouteQuery':
      var client = new SampleServiceClient(baseURI);
      return output(client.CustomRouteQuery(SampleRequest(id: '123', text: 'Abide')));

    case 'OmitMe':
      var client = new SampleServiceClient(baseURI);
      return output(client.OmitMe(SampleRequest()));

    case 'Download':
      var client = new SampleServiceClient(baseURI);
      return outputRaw(client.Download(SampleDownloadRequest(format: 'text/plain')));

    case 'DownloadResumable':
      var client = new SampleServiceClient(baseURI);
      return outputRaw(client.DownloadResumable(SampleDownloadRequest(format: 'text/plain')));

    case 'Redirect':
      var client = new SampleServiceClient(baseURI);
      return outputRaw(client.Redirect(SampleRedirectRequest()));

    case 'Authorization':
      var client = new SampleServiceClient(baseURI);
      return output(client.Authorization(SampleRequest(), authorization: 'Abide'));

    case 'AuthorizationGlobal':
      var client = new SampleServiceClient(baseURI, authorization: '12345');
      return output(client.Authorization(SampleRequest()));

    case 'AuthorizationOverride':
      var client = new SampleServiceClient(baseURI, authorization: '12345');
      return output(client.Authorization(SampleRequest(), authorization: 'Abide'));

    default:
      print('Unknown test case: "$name"');
      exit(1);
  }
}

output(Future<ModelJSON> model) async {
  try {
    var jsonString = jsonEncode(await model);
    print('OK ${jsonString}');
  }
  on SampleServiceException catch (err) {
    var message = err.message.replaceAll('"', '\'').trim();
    print('FAIL {"status":${err.status}, "message": "${message}"}');
  }
  catch (err) {
    print('FAIL {"message": "$err"}');
  }
}

outputRaw(Future<ModelStream> modelFuture) async {
  try {
    var model = (await modelFuture);
    var modelJson = model.toJson();

    // Content is still a future, so resolve it in order to print the raw data.
    modelJson['Content'] = await modelJson['Content'];
    print('OK ${jsonEncode(modelJson)}');
  }
  on SampleServiceException catch (err) {
    var message = err.message.replaceAll('"', '\'');
    print('FAIL {"status":${err.status}, "message": "${message}"}');
  }
  catch (err) {
    print('FAIL {"message": "$err"}');
  }
}
