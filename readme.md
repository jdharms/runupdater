# RunUpdater

RunUpdater is a Windows service that automatically syncs LiveSplit export data to Google Sheets. It was created specifically for The Legend of Zelda: A Link to the Past speedrunning data management.

## Features

- Runs as a Windows service
- Watches for changes to LiveSplit export Excel files
- Automatically updates Google Sheets with new data
- Logs all operations for troubleshooting

## Requirements

- Go 1.16 or higher
- Windows operating system
- LiveSplit configured to export run data as .xlsx files
- Google Sheets API credentials

## Installation

1. Clone the repository:
   ```
   git clone https://github.com/jdharms/runupdater.git
   cd runupdater
   ```

2. Build the executable:
   ```
   go build
   ```

3. Create a `config.json` file (see the Configuration section)

4. Install as a Windows service:
   ```
   sc create RunUpdater binPath= "path\to\runupdater.exe path\to\config.json"
   sc description RunUpdater "Syncs LiveSplit data to Google Sheets"
   sc start RunUpdater
   ```

## Configuration

Create a `config.json` file with the following structure:

```json
{
  "LiveSplitExportPath": "C:\\Path\\To\\LiveSplit\\Export.xlsx",
  "FileCheckInterval": 5000000000,
  "GoogleSheetsCredentialsPath": "credentials.json",
  "GoogleSheetID": "your-google-sheet-id-here",
  "RawDataSheet1Name": "Raw Data",
  "RawDataSheet2Name": "Segment History",
  "LogPath": "runupdater.log",
  "LogLevel": "info"
}
```

Notes:
- `FileCheckInterval` is in nanoseconds (5000000000 = 5 seconds)
- `GoogleSheetID` is the ID in the Google Sheets URL (the long string after `/d/` and before `/edit`)
- `GoogleSheetsCredentialsPath` should point to your Google Sheets API credentials JSON file (see below)

Here's an updated version of the "Google Sheets API Setup" section for your README.md that includes the detailed instructions about setting up a service account:

## Google Sheets API Setup

1. Go to the [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Navigate to "APIs & Services" > "Library"
4. Search for "Google Sheets API" and enable it
5. Go to "APIs & Services" > "Credentials"
6. Click "Create Credentials" and select "Service Account"
7. Fill in the service account details and click "Create"
8. For the role, select "Project" > "Editor" (or a more specific role with Sheets access)
9. Click "Done"
10. Click on the newly created service account email in the list
11. Go to the "Keys" tab
12. Click "Add Key" > "Create new key"
13. Choose "JSON" format and click "Create"
14. Save the downloaded JSON file as `credentials.json` in the same directory as your application (or update the path in your config)
15. **Important**: Open your Google Sheet in a browser
16. Click the "Share" button in the top-right corner
17. Enter the service account email address (it looks like `something@project-id.iam.gserviceaccount.com`)
18. Give it "Editor" access and click "Send" (no need to notify)
19. Make sure your config.json contains the correct GoogleSheetID (the long string in the URL between `/d/` and `/edit`)

## Development

To run the application in debug mode without installing it as a service:

```
go run . path\to\config.json
```

## License

[MIT License](LICENSE)