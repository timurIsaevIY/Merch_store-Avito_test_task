import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    stages: [
        { duration: '1m', target: 1000 },
    ],
    thresholds: {
        http_req_duration: ['p(99.99) < 50'],
        http_req_failed: ['rate<0.01'],
    },
};

const BASE_URL = 'http://localhost:8080/api';

function login(username, password) {
    const url = `${BASE_URL}/auth`;
    const payload = JSON.stringify({
        username: username,
        password: password,
    });
    const params = {
        headers: {
            'Content-Type': 'application/json',
        },
    };
    const res = http.post(url, payload, params);
    check(res, {
        'login successful': (r) => r.status === 200,
    });
    return res.headers['Access-Token'];
}

function buyItem(token, itemId) {
    const url = `${BASE_URL}/buy/${itemId}`;
    const params = {
        headers: {
            'Authorization': `Bearer ${token}`,
        },
    };
    const res = http.get(url, params);
    check(res, {
        'buy item successful': (r) => r.status === 200,
    });
}

function sendCoins(token, toUser, amount) {
    const url = `${BASE_URL}/sendCoin`;
    const payload = JSON.stringify({
        toUser: toUser,
        amount: amount,
    });
    const params = {
        headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json',
        },
    };
    const res = http.post(url, payload, params);
    check(res, {
        'send coins successful': (r) => r.status === 200,
    });
}

    function buyItem(token, itemId) {
        const url = `${BASE_URL}/buy/${itemId}`;
        const params = {
            headers: {
                'Access-Token': token,
            },
        };
        const res = http.get(url, params);
        check(res, {
            'buy item successful': (r) => r.status === 200,
        });
    }

    function sendCoins(token, toUser, amount) {
        const url = `${BASE_URL}/sendCoin`;
        const payload = JSON.stringify({
            toUser: toUser,
            amount: amount,
        });
        const params = {
            headers: {
                'Access-Token': token,
                'Content-Type': 'application/json',
            },
        };
        const res = http.post(url, payload, params);
        check(res, {
            'send coins successful': (r) => r.status === 200,
        });
    }

    const registeredUsers = [];

    export default function () {
        const iteration = __ITER;

        const username = `user_${__VU}_${iteration}`;

        const token = login(username, 'password');

        if (token) {
            registeredUsers.push(username);
        }

        buyItem(token, 1);

        if (registeredUsers.length > 1) {
            // Выбираем случайного получателя, который уже зарегистрирован (не самого себя)
            let recipient;
            do {
                recipient = registeredUsers[getRandomInt(0, registeredUsers.length - 1)];
            } while (recipient === username);

            sendCoins(token, recipient, 1);
        }

        sleep(1);
    }

function getRandomInt(min, max) {
    return Math.floor(Math.random() * (max - min + 1)) + min;
}
