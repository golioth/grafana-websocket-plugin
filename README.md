# WebSocket Data Source for Grafana

A WebSocket data source plugin for realtime data updates in [Grafana](https://grafana.com) Dashboards. 

# Plugin Purpose

This plugin allows you to connect a WebSocket endpoint as a Data Source for your Grafana dashboard. Unlike REST APIs, WebSockets automatically update as soon as new data is available from the source, without having to periodically fetch data from it.

# How it works

When a WebSocket endpoint is added as a data source in Grafana, the plugin will open the WebSocket connection with the source API and keep it open. When there is new data, the WebSocket endpoint will send it directly to the plugin through the open WebSocket connection and your dashboard will be updated automatically.

# Using the WebSocket Data Source

Here are the steps to configure and use the WebSocket Data Source Plugin in Grafana.

### Configure Data Source

1. Add the WebSocket data source plugin in Grafana

   - Click the gear icon on the left sidebar and choose "data sources"
   - Click "Add data source"
   - Scroll to the bottom and chose "WebSocket API" in the "Others" category.

2. Configure WebSocket endpoint information
   ![Grafana Websockets Configuration](https://user-images.githubusercontent.com/8334211/155154644-8925b616-a5e0-4c32-92bd-305696d0a4d1.png)

   - WebSocket Host (use this format): `wss://your-host/some/prefix-path`
   - Add Query Parameters and Custom Headers if necessary (consult your WebSocket API data source). This example shows the use of an API key.

3. Add a panel to the Grafana Dashboard to begin seeing data
   - Click `+` in the left sidebar. Choose "Dashboard" --> "Add a new panel"
   - Select `WebSocket API` as Data Source in the select drop-down
   - In the bottom left, set the "Fields" tab to `$`
   - Click the "Path" tab and set the path (if necessary) of the websocket endpoint that you want to connect
   - Above the panel, click the "Table view" toggle at the top of the window
   - Any data coming from the WebSocket Endpoint will be shown as JSON in the panel

   ![Graphana showing stream json packets](https://user-images.githubusercontent.com/8334211/150010134-4d553653-96f8-4427-a992-6812000f688f.png)

### Customizing Your Data View

Once you have confirmed that you are receiving realtime data, it can be visualized:

1. Make sure that Table View is turned off and choose any kind of compatible graphic from the upper right "Visualizations" list.

2. Choose "Last 5 minutes" from the time selection window in the upper right corner of the graph.

   ![Graphana WebSockets Graph](https://user-images.githubusercontent.com/8334211/150139309-1f5136fe-58af-425a-844e-c69d8b1a9492.png)
