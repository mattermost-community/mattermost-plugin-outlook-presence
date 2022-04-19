import {Store} from 'redux';

import {GlobalState} from 'mattermost-redux/types/store';

import manifest from './manifest';

import Constants from './constants';
import Actions from './actions';

// eslint-disable-next-line import/no-unresolved
import {PluginRegistry} from './types/mattermost-webapp';

export default class Plugin {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars, @typescript-eslint/no-empty-function
    public async initialize(registry: PluginRegistry, store: Store<GlobalState, any>) {
        registry.registerWebSocketEventHandler(Constants.STATUS_CHANGED, (event: any) => {
            store.dispatch(Actions.receivedStatusChangedEvent(event.data));
        });
    }
}

declare global {
    interface Window {
        registerPlugin(id: string, plugin: Plugin): void
    }
}

window.registerPlugin(manifest.id, new Plugin());
