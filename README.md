# WebSocket Data Source for Grafana

A WebSocket data source plugin for loading data as soon they are available at the origin [Grafana](https://grafana.com).

![Graphana showing lightDB stream json packets](https://user-images.githubusercontent.com/8334211/150010134-4d553653-96f8-4427-a992-6812000f688f.png)

# Using WebSocket Data Source

Following we will show you how to configure and use the WebSocket Data Source Plugin in your Grafana.

### Configure Data Source

1. Add The WebSocket Data Source Plugin in your Grafana

   - Click the gear icon on the left sidebar and choose "data sources"
   - Click "Add data source"
   - Scroll to the bottom and chose "WebSocket API" in the "Others" category. (If this entry is not an option, you may have a problem with allowing unsigned sources.)

2. Configure it with WebSocket Endpoint information

   ![Grafana Websockets Configuration](https://user-images.githubusercontent.com/8334211/150139294-8a374964-b9a4-40c4-8cce-7b4b32d21f69.png)

   - WebSocket Host (should use this format): `wss://your-host/some/prefix-path`
   - Add Query Parameters and Custom Headers if necessary (according the rules of the WebSocket API data source)

3. Add a panel to Grafana Dashboard and start seeing data coming in

   - Click `+` in the left sidbar. Choose "Dashboard" --> "Add a new panel"
   - In the bottom left, set the "Fields" tab to `$`
   - Click the "Path" tab and set the path (if necessary) of the websocket endpoint that you want to connect
   - Above the panel, click the "Table view" toggle at the top of the window
   - Any data coming from the WebSocket Endpoint will be shown as JSON in the panel

   ![Graphana showing stream json packets](https://user-images.githubusercontent.com/8334211/150010134-4d553653-96f8-4427-a992-6812000f688f.png)

### Sample Data Source and Customizing Your Data View

2. Make sure that Table View is turned off and choose "Time series" from the upper right "Visualizations" list.

3. Choose "Last 5 minutes" from the time selection window in the upper right corner of the graph.

   ![Graphana Websockets Graph](https://user-images.githubusercontent.com/8334211/150139309-1f5136fe-58af-425a-844e-c69d8b1a9492.png)
