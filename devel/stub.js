const http = require('http');
const url = require('url');
const querystring = require('querystring');

const sampleLogs = [
    {
        timestamp: 1573102644,
        tag: 'azure_ad.login',
        log: {
            "eventVersion": "1.0",
            "userIdentity": {
                "type": "IAMUser",
                "principalId": "EX_PRINCIPAL_ID",
                "arn": "arn:aws:iam::123456789012:user/Alice",
                "accountId": "123456789012",
                "accessKeyId": "EXAMPLE_KEY_ID",
                "userName": "Alice",
                "sessionContext": {"attributes": {
                    "mfaAuthenticated": "false",
                    "creationDate": "2014-03-06T15:15:06Z"
                }}
            },
            "eventTime": "2014-03-06T17:10:34Z",
            "eventSource": "ec2.amazonaws.com",
            "eventName": "CreateKeyPair",
            "awsRegion": "us-east-2",
            "sourceIPAddress": "72.21.198.64",
            "userAgent": "EC2ConsoleBackend, aws-sdk-java/Linux/x.xx.fleetxen Java_HotSpot(TM)_64-Bit_Server_VM/xx",
            "requestParameters": {"keyName": "mykeypair"},
            "responseElements": {
                "keyName": "mykeypair",
                "keyFingerprint": "30:1d:46:d0:5b:ad:7e:1b:b6:70:62:8b:ff:38:b5:e9:ab:5d:b8:21",
                "keyMaterial": "\u003csensitiveDataRemoved\u003e"
            }
        },
    },

    {
        timestamp: 1234456789,
        tag: 'aws.cloudtrail',
        log: {
            "eventVersion": "1.0",
            "userIdentity": {
                "type": "IAMUser",
                "principalId": "EX_PRINCIPAL_ID",
                "arn": "arn:aws:iam::123456789012:user/Alice",
                "accessKeyId": "EXAMPLE_KEY_ID",
                "accountId": "123456789012",
                "userName": "Alice"
            },
            "eventTime": "2014-03-06T21:22:54Z",
            "eventSource": "ec2.amazonaws.com",
            "eventName": "StartInstances",
            "awsRegion": "us-east-2",
            "sourceIPAddress": "205.251.233.176",
            "userAgent": "ec2-api-tools 1.6.12.2",
            "requestParameters": {"instancesSet": {"items": [{"instanceId": "i-ebeaf9e2"}]}},
            "responseElements": {"instancesSet": {"items": [{
                "instanceId": "i-ebeaf9e2",
                "currentState": {
                    "code": 0,
                    "name": "pending"
                },
                "previousState": {
                    "code": 80,
                    "name": "stopped"
                }
            }]}}
        },
    },
    {
        timestamp: 1573102344,
        tag: 'aws.cloudtrail',
        log: {
            "eventVersion": "1.0",
            "userIdentity": {
                "type": "IAMUser",
                "principalId": "EX_PRINCIPAL_ID",
                "arn": "arn:aws:iam::123456789012:user/Alice",
                "accountId": "123456789012",
                "accessKeyId": "EXAMPLE_KEY_ID",
                "userName": "Alice"
            },
            "eventTime": "2014-03-06T21:01:59Z",
            "eventSource": "ec2.amazonaws.com",
            "eventName": "StopInstances",
            "awsRegion": "us-east-2",
            "sourceIPAddress": "205.251.233.176",
            "userAgent": "ec2-api-tools 1.6.12.2",
            "requestParameters": {
                "instancesSet": {"items": [{"instanceId": "i-ebeaf9e2"}]},
                "force": false
            },
            "responseElements": {"instancesSet": {"items": [{
                "instanceId": "i-ebeaf9e2",
                "currentState": {
                    "code": 64,
                    "name": "stopping"
                },
                "previousState": {
                    "code": 16,
                    "name": "running"
                }
            }]}}
        },
    },
    {
        timestamp: 1573102644,
        tag: 'aws.cloudtrail',
        log: {
            "eventVersion": "1.0",
            "userIdentity": {
                "type": "IAMUser",
                "principalId": "EX_PRINCIPAL_ID",
                "arn": "arn:aws:iam::123456789012:user/Alice",
                "accountId": "123456789012",
                "accessKeyId": "EXAMPLE_KEY_ID",
                "userName": "Alice",
                "sessionContext": {"attributes": {
                    "mfaAuthenticated": "false",
                    "creationDate": "2014-03-06T15:15:06Z"
                }}
            },
            "eventTime": "2014-03-06T17:10:34Z",
            "eventSource": "ec2.amazonaws.com",
            "eventName": "CreateKeyPair",
            "awsRegion": "us-east-2",
            "sourceIPAddress": "72.21.198.64",
            "userAgent": "EC2ConsoleBackend, aws-sdk-java/Linux/x.xx.fleetxen Java_HotSpot(TM)_64-Bit_Server_VM/xx",
            "requestParameters": {"keyName": "mykeypair"},
            "responseElements": {
                "keyName": "mykeypair",
                "keyFingerprint": "30:1d:46:d0:5b:ad:7e:1b:b6:70:62:8b:ff:38:b5:e9:ab:5d:b8:21",
                "keyMaterial": "\u003csensitiveDataRemoved\u003e"
            }
        },
    },
    {
        timestamp: 1573102644,
        tag: 'sample.testdata1',
        log: {
            "longvalue1": "Weobservetodaynotavictoryofpartybutacelebrationoffreedomsymbolizinganendaswellasabeginningsignifyingrenewalaswellaschange.ForIhaveswornbeforeyouandAlmightyGodthesamesolemnoathourforbearsprescribednearlyacenturyandthreequartersago.",
            "longvalue2": "Wedarenotforgettodaythatwearetheheirsofthatfirstrevolution.Letthewordgoforthfromthistimeandplace,tofriendandfoealike,thatthetorchhasbeenpassedtoanewgenerationofAmericansborninthiscentury,temperedbywar,disciplinedbyahardandbitterpeace,proudofourancientheritageandunwillingtowitnessorpermittheslowundoingofthosehumanrightstowhichthisnationhasalwaysbeencommitted,andtowhichwearecommittedtodayathomeandaroundtheworld.",
            "longvalue3": "Leteverynationknow,whetheritwishesuswellorill,thatweshallpayanyprice,bearanyburden,meetanyhardship,supportanyfriend,opposeanyfoetoassurethesurvivalandthesuccessofliberty.            ",
            "looooooooooooooooooooooooooooooooooonnnnnnnnnnnnnnnnnnnnnnnngggggggggggggggggggggggggggggggg value": "not long",
        },
    },
    {
        timestamp: 1573102644,
        tag: 'sample.testdata2',
        log: {
            "message": "XYGAOISF: warning: header Subject: I Love You from unknown[10.1.2.3]; from=<noreply@example.com> to=<mizutani@example.com> proto=ESMTP helo=<mx.example.com>",
        },
    },
];

const server = http.createServer((req, res) => {
    URL = url.parse(req.url);
    const qs = querystring.parse(URL.query)
    console.log(req.method, URL, req.headers, qs);

    const offset = (qs.offset === undefined ? 0 : parseInt(qs.offset));

    if (req.method === 'POST') {
        const result = {'query_id': 'abcdefg'};
        res.writeHead(200, { 'Content-Type': 'application/json'});
        res.write(JSON.stringify(result));
        res.end();
    } else if (req.method === 'GET') {
        const result = {
            metadata:{
                "status": "SUCCEEDED",
                "total": 35,
                "offset": offset,
                "limit": 10,
                "submitted_time": "2019-11-23T13:30:48Z",
                "elapsed_seconds": 17
            },
            logs: sampleLogs,
        };

        res.writeHead(200, { 'Content-Type': 'application/json'});
        res.write(JSON.stringify(result));
        res.end();
    } else {
        res.end();
    }
});

server.on('clientError', (err, socket) => {
  socket.end('HTTP/1.1 400 Bad Request\r\n\r\n');
});

server.listen(8000);