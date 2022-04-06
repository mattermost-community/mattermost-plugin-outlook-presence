export interface PluginRegistry {
    registerPostTypeComponent(typeName: string, component: React.ElementType)
    registerWebSocketEventHandler(event: string, handler: (msg: any) => void)

    // Add more if needed from https://developers.mattermost.com/extend/plugins/webapp/reference
}
