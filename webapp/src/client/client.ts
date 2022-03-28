import axios, {AxiosInstance, AxiosResponse} from 'axios';

import Constants from '../constants';

interface StatusChangedEvent {
    userId: string;
    status: string;
}

export default class Client {
    url: URL;
    baseUrl: string;
    pluginUrl: string;
    pluginApiUrl: string;
    client: AxiosInstance;

    constructor() {
        this.url = new URL(window.location.href);
        this.baseUrl = `${this.url.protocol}//${this.url.host}`;
        this.pluginUrl = `${this.baseUrl}/plugins/${Constants.PLUGIN_NAME}`;
        this.pluginApiUrl = `${this.pluginUrl}/api/v1`;
        this.client = axios.create({
            baseURL: this.pluginApiUrl,
            headers: {
                'Content-Type': 'application/json',
            },
        });
    }

    postStatusChanged = (data: StatusChangedEvent) => {
        const {userId, status} = data;
        return this.doPost(`${this.pluginApiUrl}/status/publish`, {
            user_id: userId,
            status,
        });
    }

    doPost = async (url: string, body: any, headers: any = {}): Promise<AxiosResponse<any, any>> => {
        return this.client.post(url, body, {headers});
    };
}
