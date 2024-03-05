/* global require, process */

// We don't want node to include output regarding fetch being experimental... just use it.
process.removeAllListeners('warning')

// These should be copied into place when we run 'make generate' for testing.
// const {SampleServiceClient} = require('./sample_service.gen.client');
import {SampleServiceClient} from './sample_service.gen.client.js';

async function main() {
    if (process.argv.length < 4) {
        console.info("Usage:");
        console.info("  node run_client.js {ADDRESS} {TEST_CASE}");
        console.info("");
        console.info("Example:");
        console.info("  node run_client.js localhost:9000 DownloadResumable");
        console.info("  node run_client.js localhost:9001 Authorization");
    }

    const baseURI = 'http://' + process.argv[2];
    const testCase = process.argv[3];

    switch (testCase) {
        case 'NotConnected': {
            const client = new SampleServiceClient(baseURI);
            return output(client.Defaults({Text: 'Hello'}));
        }
        case 'BadFetch': {
            const client = new SampleServiceClient(baseURI, {fetch: 'fart'});
            return output(client.Defaults({Text: 'Hello'}));
        }
        case 'Defaults': {
            const client = new SampleServiceClient(baseURI);
            return output(client.Defaults({Text: 'Hello'}));
        }
        case 'ComplexValues': {
            const client = new SampleServiceClient(baseURI);
            return output(client.ComplexValues({
                InUser: {
                    ID: '123',
                    Name: 'Dude',
                    Age: 47,
                    Attention: 1000000,
                    AttentionString: '4m2s',
                    Digits: '555-1234',
                    MarshalToString: "home@string.com,work@string.com",
                    MarshalToObject: { H: "home@object.com", W: "work@object.com" },
                },
                InFlag: true,
                InFloat: 3.14,
                InTime: '2022-12-05T17:47:12+00:00',
                InTimePtr: '2020-11-06T17:47:12+00:00',
            }));
        }
        case 'ComplexValuesPath': {
            const client = new SampleServiceClient(baseURI);
            return output(client.ComplexValuesPath({
                InUser: {
                    ID: '123',
                    Name: 'Dude',
                    Age: 47,
                    Attention: 1000000,
                    AttentionString: '4m2s',
                    Digits: '555-1234',
                    MarshalToString: "home@string.com,work@string.com",
                    MarshalToObject: { H: "home@object.com", W: "work@object.com" },
                },
                InFlag: true,
                InFloat: 3.14,
                InTime: '2022-12-05T17:47:12+00:00',
                InTimePtr: '2020-11-06T17:47:12+00:00',
            }));
        }
        case 'Fail4XX': {
            const client = new SampleServiceClient(baseURI);
            return output(client.Fail4XX({}));
        }
        case 'Fail5XX': {
            const client = new SampleServiceClient(baseURI);
            return output(client.Fail5XX({}));
        }
        case 'CustomRoute': {
            const client = new SampleServiceClient(baseURI);
            return output(client.CustomRoute({ID: '123', Text: 'Abide'}));
        }
        case 'CustomRouteQuery': {
            const client = new SampleServiceClient(baseURI);
            return output(client.CustomRouteQuery({ID: '456', Text: 'Abide'}));
        }
        case 'CustomRouteBody': {
            const client = new SampleServiceClient(baseURI);
            return output(client.CustomRouteBody({ID: '789', Text: 'Abide'}));
        }
        case 'OmitMe': {
            const client = new SampleServiceClient(baseURI);
            return output(client.OmitMe({}));
        }
        case 'Download': {
            const client = new SampleServiceClient(baseURI);
            return output(client.Download({Format: 'text/plain'}));
        }
        case 'DownloadResumable': {
            const client = new SampleServiceClient(baseURI);
            return output(client.DownloadResumable({Format: 'text/plain'}));
        }
        case 'Redirect': {
            const client = new SampleServiceClient(baseURI);
            return output(client.Redirect({}));
        }
        case 'Authorization': {
            const client = new SampleServiceClient(baseURI);
            return output(client.Authorization({}, {authorization: 'Abide'}));
        }
        case 'AuthorizationGlobal': {
            const client = new SampleServiceClient(baseURI, {authorization: '12345'});
            return output(client.Authorization({}));
        }
        case 'AuthorizationOverride': {
            const client = new SampleServiceClient(baseURI, {authorization: '12345'});
            return output(client.Authorization({}, {authorization: 'Abide'}));
        }
    }
}

async function output(responseFuture) {
    try {
        const value = await responseFuture;
        const content = value && value['Content'];

        // Duck typing to determine if we received a raw Blob back instead of decoded JSON. Our
        // raw stream responses all return byte streams of text, so just flatten to that value.
        if (content && content['arrayBuffer'] && content['text']) {
            value.Content = await value.Content.text();
        }
        console.info('OK ' + JSON.stringify(await value));
    }
    catch (e) {
        const failure = await e;
        const failureJSON = typeof failure === 'string'
            ? JSON.stringify({message: failure})
            : JSON.stringify(failure);

        console.info('FAIL ' + failureJSON);
    }
}

main().then().catch((e) => console.info('FAILURE:' + e));
