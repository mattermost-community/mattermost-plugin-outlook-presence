# Mattermost Plugin Outlook Presence

This plugin is used for getting status updates for users. It adds an additional websocket endpoint to the Mattermost server which can be connected to without an active user's valid token. It also exposes a new paginated API to get the statuses of all users along with their emails. The purpose of this plugin is to provide status updates to Microsoft Outlook through the use of an intermediary [IM app](https://github.com/mattermost/mattermost-outlook-presence-provider). To know more about how to integrate an application with Outlook, you can read the [official docs](https://docs.microsoft.com/en-us/office/client-developer/shared/integrating-im-applications-with-office). (This plugin can be used with any third-party application that wants to subscribe to Mattermost users' status updates.)

## Prerequisite

This plugin works with a companion Windows binary which integrates with Outlook on each user's desktop. Learn more about the **Mattermost Outlook Presence Provider** [here](https://github.com/mattermost/mattermost-outlook-presence-provider).

## Installation

1. Go to the [releases page of this GitHub repository](https://github.com/mattermost/mattermost-plugin-outlook-presence/releases) and download the latest release for your Mattermost server.
1. Upload this file in the Mattermost **System Console > Plugins > Management** page to install the plugin. To learn more about how to upload a plugin, [see the documentation](https://docs.mattermost.com/administration/plugins.html#custom-plugins).
1. After installing the plugin, you should go to the plugin's settings in System Console and generate a Webhook Secret (more about this below).

## Configuration

This plugin contains some configuration settings that you must set. You can do that by going to **System Console > Plugins > Outlook Presence**. You can know more about these settings below:

- **Webhook Secret**:
  Setting a webhook secret allows you to ensure that the requests sent to the payload URL are from the Office IM app (or any other client app), and is used with every request that is made from the IM app to Mattermost.

 - **Status response page size**
  This setting is for the paginated APIs exposed by the plugin. It basically denotes the number of statuses to return on a single page of the API request.

## Features
This plugin adds two endpoints to the Mattermost server and both require authentication using the webhook secret in the plugin configuration settings.

- **GetStatusForAllUsers endpoint**: `/status` is the endpoint which can be used to get the statuses for all **active** users present in Mattermost. The request must contain the `webhook secret` in a query param called `secret` or in form data. It accepts another query param called `page` whose default value is `0`. If a page does not contain any users, then the endpoint returns an empty array. Also, if there's no record of a user's status in the Mattermost database (in the case of bots and users who have just signed up), then this endpoint returns their status as "offline".

- **Websocket endpoint**: `/ws` is the endpoint through which you can connect to the websocket. This plugin adds server logs whenever a new client is connected/disconnected along with the current size of the websocket connection pool. This endpoint also requires the `secret` query param for authentication.

You can make a request to both these endpoints using the base url as - 
```
{MATTERMOST_SERVER_URL}/plugins/com.mattermost.outlook-presence/api/v1
```

## Building the plugin

- Make sure you have following components installed:
    - Go - v1.16 - [Getting Started](https://golang.org/doc/install)
      > **Note:** If you have installed Go to a custom location, make sure the `$GOROOT` variable is set properly. Refer [Installing to a custom location](https://golang.org/doc/install#install).
    - NodeJS - v14.17 and NPM - [Downloading and installing Node.js and npm](https://docs.npmjs.com/getting-started/installing-node).
    - Make

- Note that this project uses [Go modules](https://github.com/golang/go/wiki/Modules). Be sure to locate the project outside of `$GOPATH`.
To learn more about plugins, see [plugin documentation](https://developers.mattermost.com/extend/plugins/).

- Build your plugin:
    ```
    make dist
    ```

- This will produce a single plugin file (with support for multiple architectures) for upload to your Mattermost server:
    ```
    dist/com.mattermost.outlook-presence-x.y.z.tar.gz
    ```

## Development

To avoid having to manually install your plugin, build and deploy your plugin using one of the following options. In order for the below options to work, you must first enable plugin uploads via your config.json or API and restart Mattermost.

```json
    "PluginSettings" : {
        ...
        "EnableUploads" : true
    }
```

### Deploying with Local Mode

If your Mattermost server is running locally, you can enable [local mode](https://docs.mattermost.com/administration/mmctl-cli-tool.html#local-mode) to streamline deploying your plugin. Edit your server configuration as follows:

```json
{
    "ServiceSettings": {
        ...
        "EnableLocalMode": true,
        "LocalModeSocketLocation": "/var/tmp/mattermost_local.socket"
    },
}
```

and then deploy your plugin:
```
make deploy
```

You may also customize the Unix socket path:
```
export MM_LOCALSOCKETPATH=/var/tmp/alternate_local.socket
make deploy
```

### Deploying with credentials

Alternatively, you can authenticate with the server's API with credentials:
```
export MM_SERVICESETTINGS_SITEURL=http://localhost:8065
export MM_ADMIN_USERNAME=admin
export MM_ADMIN_PASSWORD=password
make deploy
```

or with a [personal access token](https://docs.mattermost.com/developer/personal-access-tokens.html):
```
export MM_SERVICESETTINGS_SITEURL=http://localhost:8065
export MM_ADMIN_TOKEN=j44acwd8obn78cdcx7koid4jkr
make deploy
```

Made with &#9829; by [Brightscout](https://www.brightscout.com)
