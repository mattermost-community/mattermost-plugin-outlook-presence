import {ActionFunc, DispatchFunc} from 'mattermost-redux/types/actions';
import {logError} from 'mattermost-redux/actions/errors';

import Client from 'client';

const receivedStatusChangedEvent = (data: any): ActionFunc => {
    return async (dispatch: DispatchFunc) => {
        Client.postStatusChanged({
            userId: data.user_id,
            status: data.status,
        }).catch((err: any) => {
            dispatch(logError(err));
        });

        return {data: true};
    };
};

export default {
    receivedStatusChangedEvent,
};
