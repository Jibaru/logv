# logv

**logv** is an interactive log viewer written in Go that enhances your logging experience by providing real-time filtering, JSON formatting, and syntax highlighting. It reads log lines from standard input, allowing you to pipe logs into the application, and then displays them in an intuitive user interface. This tool is ideal for developers who need to monitor and debug logs efficiently in local environment.

## Features

- **Interactive Log Viewing:** Display logs in an interactive list that updates in real time.
- **Live Search Filtering:** Dynamically filter log entries using a search input.
- **JSON Formatting:** Automatically detect and prettify JSON logs, making them easier to read.
- **Syntax Highlighting:** Highlights keys and string values within JSON logs for better readability.
- **Cross-Platform Support:** Works on Unix-like systems and Windows. Uses `/dev/tty` on Unix and `CONIN$` on Windows for interactive input.
- **Keyboard Shortcuts:** Easily exit the application using the `q` key.

## Installation

To install **logv**, ensure you have [Go installed](https://golang.org/doc/install) and then run:

```bash
go install github.com/Jibaru/logv
```

This command downloads the source code, compiles the executable, and installs it into your `$GOPATH/bin`. Make sure that your Go bin directory is included in your system's PATH to execute **logv** from anywhere.

## Usage

**logv** is designed to be used as part of a logging pipeline. It reads log data from standard input, processes it, and provides an interactive UI for browsing through log entries.

### Basic Example

Suppose you have a Python script, `generate_logs.py`, that continuously generates log entries. You can pipe its output into **logv** using the following command:

```bash
python generate_logs.py | logv.exe
```

In this example:

- `python generate_logs.py` is a hypothetical script that produces log output.
- The pipe (`|`) redirects the output of the Python script into **logv**.
- `logv.exe` starts the log viewer application (on Windows the executable might be named `logv.exe`, whereas on Unix-like systems it will simply be `logv`).

Once started, **logv** opens an interactive UI:

- Use the arrow keys or mouse to navigate through the list of log entries.
- Type in the search box at the top to filter the logs.
- Press `Enter` on a log entry to view it in a modal window with formatted and highlighted JSON.
- Press `q` to quit the application.

## Detailed Walkthrough

### Log Input Processing

**logv** is designed to read from standard input. Whether logs are coming from a file, a pipe, or another process, **logv** captures each line, stores it in memory, and immediately updates the UI if the log line matches the current search filter.

### JSON Formatting & Highlighting

If a log entry is a valid JSON string, **logv** will:

- **Prettify** the JSON by adding indentation.
- **Highlight** JSON keys in yellow and string values in green.
  This allows for easy visual parsing of complex log structures.

### Interactive User Interface

Using the [tview](https://github.com/rivo/tview) and [tcell](https://github.com/gdamore/tcell) libraries, **logv** creates a rich, interactive terminal-based UI that:

- Displays log entries in a list.
- Provides an input field for real-time search filtering.
- Opens a modal window for detailed log inspection when a log entry is selected.
- Supports mouse input and intuitive keyboard controls.

## License

**logv** is released under the MIT License. See the [LICENSE](LICENSE) file for details.
