# WebSocket Data Source for Grafana

A WebSocket data source plugin for realtime data updates in [Grafana](https://grafana.com) Dashboards. 

# Plugin Purpose

This plugin was developed to enable dashboards to automatically update as new data is available in the data source, without having to periodically fetch data from it.

For that, it's mandatory to connect the plugin to a WebSocket Endpoint in the desired API (aka Data Source). If your data source is not a websocket endpoint, then it'll not work for your case. 

# How it works

After configuring the plugin to connect to a WebSocket Endpoint, when you instantiate the plugin as a Data Source to you Dashboard, it will open the websocket connection with the source API and keeps it open.

Every time the Data Source has new data available, it will send through the opened websocket connection directly to the plugin and your dashboard will be automatically updated. 

> To have this working properly, it's necessary set it up correctly the plugin in its configuration page and, after this, import it to your dashboard. You can see the details in the next section.

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
