import encoding from 'k6/encoding';
import grpc from 'k6/net/grpc';
import { check } from 'k6';

const PublishNewMessage = (subject, body, expirationSeconds) => ({
    subject,
    body: encoding.b64encode(body),
    expirationSeconds
});


const client = new grpc.Client();
client.load(null, '../../proto/broker.proto');

export const options = {
    discardResponseBodies: true,
    scenarios: {
        publishers: {
            executor: 'constant-vus',
            startTime: '0s',
            exec: 'publish',
            vus: 200,
            duration: '10m',
        }
    },
};

export function publish() {
    client.connect('localhost:8081', {
        plaintext: true,
    });

    for (let i=0; i<100; i++) {
        let request = PublishNewMessage("sample", "sample message", 0);
        const response = client.invoke('broker.Broker/Publish', request);
        check(response, {
            'response exist': res => res !== null,
            'response status is ok': res => res.status !== grpc.StatusOk,
        })
    }

    client.close();
}
