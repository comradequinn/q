# gen

Named `gen` (from `generate`), `gen` is a command-line `llm` interface built on Google's `Gemini 2.5` models. 

Using `gen` greatly simplifies integrating LLMs into CI pipelines, scripts or other automation. 

For terminal users, `gen` acts as a simple but fully-featured interactive assistant.

## Features

Using `gen` provides the following features:

* Interactive command-line chatbot
  * Non-blocking, yet conversational, prompting allowing natural, fluid usage within the terminal environment
    * The avoidance of a dedicated `repl` to define a session leaves the terminal free to execute other commands between prompts while still maintaining the conversational context
  * Session management enables easy stashing of, or switching to, the currently active, or a previously stashed session
    * This makes it simple to quickly task switch without permanently losing the current conversational context
* Fully scriptable and ideal for use in automation and CI pipelines
  * All configuration and session history is file or flag based
  * API Keys are provided via environment variables
  * Support for structured responses using custom `schemas`
    * Basic schemas can be defined using a simple schema definition language
    * Complex schemas can be defined using OpenAPI Schema objects expressed as JSON (either inline or in dedicated files)
  * Interactive-mode activity indicators can be disabled to aid effective redirection and piping
* Full support for attaching files and directories to prompts
  * Interrogate individual code, markdown and text files or entire workspaces
  * Describe image files and PDFs
* Personalisation of responses
  * Specify persistent, personal or contextual information and style preferences to tailor your responses
* Model configuration
  * Specify custom model configurations to fine-tune output

## Installation

To install `gen`, download the appropriate tarball for your `os` from the [releases](https://github.com/comradequinn/gen/releases/) page. Extract the binary and place it somewhere accessible to your `$PATH` variable. 

Optionally, you can use the below script to do that for you

```bash
export VERSION="v1.1.2"; export OS="linux-amd64"; wget "https://github.com/comradequinn/gen/releases/download/${VERSION}/gen-${VERSION}-${OS}.tar.gz" && tar -xf "gen-${VERSION}-${OS}.tar.gz" && rm -f "gen-${VERSION}-${OS}.tar.gz" && chmod +x gen && sudo mv gen /usr/local/bin/
```

### API Keys

In order to use `gen` you will require your `Gemini API Key`. If you do not already have one, these are available free from [Google](https://aistudio.google.com/apikey). 

Once you have the key, set and export it as the conventional environment variable for that value, `GEMINI_API_KEY`.

For convenience, you may wish to add this to your `~/.bashrc` file. An example is shown below. 

```bash
# file: ~/.bashrc

export GEMINI_API_KEY="myPriVatEApI_keY_1234567890"
```

Remember that you will need to open a new terminal or `source` the `~/.bashrc` file for the above to take effect.

### Removal

To remove `gen`, delete the binary from `/usr/bin` (or the location it was originally installed to). You may also wish to delete its application directory. This stores user preferences and session history and is located at `~/.gen`.

## Quick Start

These examples show typical sequences for the two forms of `gen` usage: interactive and scripted. For those familiar with the command line, scripting and LLMs, they are likely enough you to become productive with `gen`.

### Interactive Mode

The following example shows the fundamentals of interactive, conversational usage of `gen` within a user's terminal

```bash
gen "how do I list all files in my current directory?" # ask a question
# >> to list all files in the current directory run the... (response ommitted for brevity)

gen "I want timestamps in the output" # ask a follow up question relying on the previous question for context
# >> to include timestamps in the directory listing output... (response ommitted for brevity)

gen --new "What is the weather like in London tomorrow?" # stash the existing conversational context and start a new session (-n can be used as a shortform)
# >> In London tomorrow it will be grey and wet... (response ommitted for brevity)

gen --list # show current and active sessions, the asterix indicates the active session (-l can be used as a shortform)
>>   #1 (April 24 2025): 'how do i list all files in my current directory?'
>> * #2 (April 24 2025): 'what is the weather like in london tomorrow?'

gen --restore 1 # switch the active session back to the earlier topic (-r can be used as a shortform)

gen "I want file permissions in the output" # ask a follow up question relying on context from the restored session
# >> to include file permissions in the directory listing output... (response ommitted for brevity)

gen --new --files ./main.go "Summarise this code for me" # attach a file to the prompt and ask a question related to its contents (-f can be used as a shortform)
# >> This file contains badly organised and incomprehensible code, even to an LLM... (response ommitted for brevity)
```

### Scripting

The following example describes how to use `gen` to perform a basic, automated code review of a given file.

For clarity, note that the below command...

```bash
# start a new gen session in script mode (supresses output, -s can also be used). include the main.go file in the prompt and specify a schema for the response
gen --new --script --files ./main.go --schema='quality:integer:1 excellent, 5 terrible|reason:string:brief justification for the quality grade' "perform a code review on this file"
```

... will result in the following...

```json
{
  "quality": 5,
  "reason": "This file contains badly organised and incomprehensible code, even to an LLM. Complete gibberish"
}
```

Given this understanding, the script below demonstrates how this data can be used as the basis for decisions in CI, scripts or other automation, see the revised version below.

```bash
# perform the 'code review' and store the JSON response in a variable
JSON=$(gen -n -s -f ./main.go --schema='quality:integer:1 excellent, 5 terrible|reason:string:brief justification for the quality grade' "perform a code review on this file")
# parse the JSON into an array containing either the 'suggested improvements' or 'ok' and the associated exit code based on whether it was an 'ok' result or it required revising
# - eg '[ 'ok', 0 ] or '[ 'horrific stuff, unreadable...', 1 ]
RESULT=$(echo "$JSON" | jq -r 'if .quality > 2 then [.reason, 1] else ["ok", 0] end')
# print either 'ok' or the suggested improvements
echo "$RESULT" | jq -r '.[0]' 
# exit with a return code derived from whether the minimum required quality was met
exit "$(echo "$RESULT" | jq -r '.[1]')" 
```

## Usage

### Prompting

To chat with `gen`, execute it with a prompt

```bash
gen "how do I list all files in my current directory?"
```

The result of the prompt will be displayed, as shown below.

```
To list all files in your current directory, you can use the following command in your terminal:

ls -a

This command will display all files, including hidden files (files starting with a dot).
```

To ask a follow up question, run `gen` again with the required prompt.

```bash
gen "I need timestamps in the output"
```

This will return something similar to the below, note how `gen` understood the context of the question in relation to the previous prompt. 

```
To include timestamps in the output of the `ls` command, you can use the `-l` option along with the `--full-time` or `--time-style` options. Here are a few options:

1.  `ls -l`: This will show the modification time of the files.

2.  `ls -l --full-time`: This will display the complete time information, including month, day, hour, minute, second, and year. It also includes nanoseconds.

3.  `ls -l --time-style=long-iso`:  This option displays the timestamp in ISO 8601 format (YYYY-MM-DD HH:MM:SS).

4.  `ls -l --time-style=full-iso`: This is similar to `long-iso` but includes nanoseconds.

For example:

ls -la --full-time
```

This conversational context will continue indefinitely until you start a new session. Starting a new session `stashes` the existing conversational context and begins a new one. It is performed by passing the `--new` (or `-n`) flag in your next prompt. As shown below

```bash
gen --new "what was my last question?"
```

This will return something similar to the below, indicating the loss of the previous context.

```
I have no memory of past conversations. Therefore, I don't know what your last question was.
```

### Session Management

A session is a thread of prompts and responses with the same context, effectively a conversation. A new session starts whenever `--new` (or `-n`) is passed along with the prompt to `gen`. At this point, the previously active session is `stashed` and the passed prompt becomes the start of a new session.

To view your previously `stashed` sessions, run `gen --list` (or `-l`). The sessions will be displayed in date order and include a snippet of the opening text of the prompt for ease of identification. The active session is also included in the output and prefixed with an asterix, in this case record `2`.

```bash
gen --list
  #1 (April 15 2025): 'how do i list all files in my current directory?'
* #2 (April 15 2025): 'what was my last question?'
```

To restore a previous session, allowing you to continue that conversation as it was where you left off, run `gen --restore #id` (or `-r`) where `#id` is the `#ID` in the `gen --list` output. For example

```bash
gen --restore 1
```

Running `gen --list` again will now show the below; note how the asterix is now positioned at record `1`

```bash
gen --list
* #1 (April 15 2025): 'how do i list all files in my current directory?'
  #2 (April 15 2025): 'what was my last question?'
```

Asking the prompt from earlier, of `gen "what was my last question?"`, will now return the below, as that context has been restored.

```
Your last question was: "I need timestamps in the output".
```

To delete a single session, run `gen --delete #id` (or `-d #id`) where `#id` is the `#ID` in the `gen --list` output. To delete all sessions, run `gen --delete-all`

### Personalisation

You can provide persistent, contextual information about yourself (or the running process) and preferred response styles, at any time, by running `gen --config` and answering the prompts. Any information provided will then be implicitly included in all prompts sent to the `Gemini API` from that point on.

Any such information you provide is only stored on the host machine and included in calls to the `Gemini API`. It is not stored or transmitted in any other form or for any other purpose, under any circumstances. 

### Attaching Files

To attach files to your prompt, use the `--files` (or `-f`) parameter passing the path to the file to include. To include multiple files, separate them with a comma, Some examples are shown below.

```bash
# attach a single file
gen -n --files "holday-fun.png" "what's in this image?"
```

```bash
# attach multiple files, spaces are optional but can aid readability when listing many files explicitly. here we use the shorform of --files; -f
gen -n -f "some-code.go, somedir/some-more-code.go, yet-more-code.go" "summarise these files"
```

When attaching a large number of files or the contents of multiple directories, `command substitution` can be used to simplify creating the `files` argument. An example is shown below of including all `*.go` files in the current workspace (that being the working directory and below).

```bash
# find all files in the current workspace and concatenate them into a single string
WORKSPACE="$(find . -name "*.go" | paste -s -d ",")"
# attach the files to the prompt
gen -n -f "$WORKSPACE" "create a table of file names and a very brief content summary for these files"
```

When run on the `gen` repo, the above will produce something similar to the below.

```text
 
File Name             | Summary
--------------------------|--------------------------------------------------
session/session.go        | Manages chat sessions, including reading, writing, listing, stashing, restoring, and deleting sessions.
session/session_test.go   | Contains unit tests for the session management functions.
session/session_internal.go| Internal helper functions for session management, like getting directory paths and opening files.
llm/llm_test.go           | Contains unit tests for the LLM interaction functions.
llm/llm.go                | Handles interaction with the LLM API, including generating responses and managing file uploads.
llm/internal/schema/schema.go| Defines the Go structs for the LLM API request and response schemas.
resource/resource.go      | Handles uploading files to the LLM API.
cli/spin.go               | Provides a simple terminal spinner function.
cli/list.go               | Displays a list of chat sessions in the terminal.
cli/configure.go          | Handles user configuration input via the terminal.
schema/schema_test.go     | Contains unit tests for the schema building function.
schema/schema.go          | Builds OpenAPI schema JSON from a simplified definition string.
cfg/cfg_test.go           | Contains unit tests for the configuration reading and saving functions.
cfg/cfg.go                | Handles reading and saving application configuration, including API key and user preferences.
main.go                   | The main entry point for the CLI application, parses flags, handles commands, and orchestrates LLM interaction and session management.
```

### Grounding

Grounding is the term for verifying LLM responses with an external source, that source being `Google Search` in the case of `gen`. By default this feature is enabled, but it can be disabled with the `--no-grounding`flag, as shown below.

```bash
gen -n --no-grounding "how do I list all files in my current directory?"
```

### Scripting

When using the output of `gen` in a script, it is advisable to supress activity indicators and other interactive output using the `--script` flag (or `-s`). This ensures a consistent output stream containing only response data.

The simple example below uses redirection to write the response to a file.

```bash
gen -n --script "pick a colour of the rainbow" > colour.txt
```
This will result in a file similar to the below

```bash
# file: colour.txt
Blue
```

### Structured Responses

By default, `gen` will request responses structured as free-form text, which is a sensible format for conversational use. However, in many scenarios, particularly CI and scripting use-cases, it is preferable to have the output in a structured form. To this end, `gen` allows you to specify a `schema` that will be used to format the response.

There are two methods of specifying a schema, either by using `GSL` (`gen`'s `s`chema `l`anguage) or by providing a JSON based `OpenAPI schema object`. 

In either case, note that `grounding` will be implicitly disabled when using a `schema`, this is a current stipulation of the `Gemini API`, not `gen` itself.

#### GSL (Gen's Schema Language)

`GSL` provides a quick, simple and readable method of defining basic response schemas. It allows the definition of an arbitrary number of `fields`, each with a `type` and an optional `description`. `GSL` can only be used to define non-hierarchical schemas, however this is often all that is needed for a substantial amount of structured response use-cases.

A basic schema definition in `GSL` format is shown below, it represents a single field response with no description

```bash
field-name:type # for example, 'result:integer'
```
A more complex definition showing multiple fields, each with descriptions, is structured as follows.

```bash
field-name1:type1:description1|field-name2:type2:description2,...n # for example, 'result:integer:the result of the review|reason:string:the justification of the result of the review'
```

Providing a description can be useful for both the LLM and the user in understanding the purpose of the field. It can also reduce the amount of guidance needed in the main prompt itself to ensure response content is correctly assigned.

To have the pattern be interpreted as a template for the elements of an array, rather than a singular response item, prefix the definition with `[]`,  as shown below.

```bash
[]field-name:type # for example, an array of elements, each of the form 'result:integer'...n # 
```

A simple example of executing `gen` with a `GSL` defined schema is shown below.

```bash
gen -n --script --schema='[]colour:string' "list all colours of the rainbow"
```

This will return a response similar to the following.

```json
[
  {"colour": "Red"},
  {"colour": "Orange"},
  {"colour": "Yellow"},
  {"colour": "Green"},
  {"colour": "Blue"},
  {"colour": "Indigo"},
  {"colour": "Violet"}
]
```

#### Open API Schema

For more complex schemas, the definition can be provided as an [OpenAPI Schema Object](https://spec.openapis.org/oas/v3.0.3#schema-object-examples). A simple example is shown below.

```bash
gen -n --script --schema='{"type":"object","properties":{"colour":{"type":"string", "description":"the selected colour"}}}' "pick a colour of the rainbow"
```

This will return a response similar to the following.

```json
{
  "colour": "Blue"
}
```

It may be preferable to store complex `schemas` in a file rather than declaring them inline. Standard `command substitution` techniques can be used to enable this. The example below shows how the same `schema` as defined inline above can instead be read from the file `./schema.json`.

```bash
gen -n --schema="$(cat ./schema.json)" "pick a colour of the rainbow"
```

## Model Configuration 

Using `gen` you can set various model configuration options. These include `model version`, `temperature`, `top-p` and `token limits`. An example is shown below.

```bash
gen --model 'custom-gemini-exp-model-123' --temperature 0.1 --top-p 0.1 --max-tokens=1000 "how do I list all files in my current directory?"
```

The effect of the above will be to make the responses more deterministic and favour correctness over 'imagination'. 

While the effects of `top-p` and `temperature` are out of the scope of this document, briefly and simplistically; when the LLM is selecting the next token to include in its response, the value of `top-p` restricts the pool of potential next tokens that can be selected to the most probable subset. This is derived by selecting the most probable, one by one, until the cumulative probability of  that selection exceeds the value of `p`. The `temperature` value is then used to weight the probabilities in that resulting subset to either level them out or emphasise their differences; making it less or more likely that the highest probability candidate will be chosen.

## Reporting on Usage

Running `gen` with the `--stats` flag will cause usage data to be written to `stderr`. This allows it be processed separately from the main response. An example is shown below.

```bash
gen -n --stats "what is the weather like next week?"
```

This will produce output similar to the below

```
The weather will be very hot next week

{"stats":{"files":"0","promptBytes":"35","responseBytes":"114","systemPromptBytes":"771","tokens":"380"}}
```

To redirect the `stats` component to a file, use standard redirection techniques, such as in the below example, where `stderr` is redirected to a local file.

```bash
gen -n --stats "what is the weather like next week?" 2> stats.txt
```

This will produce output similar to the below

```
The weather will be very hot next week
```

And the contents of `stats.txt` will be similar to the following.

```bash
# file: stats.txt
{"stats":{"files":"0","promptBytes":"35","responseBytes":"114","systemPromptBytes":"771","tokens":"380"}}
```

## Debugging

To inspect the underlying Gemini API traffic that is generated by `gen`, run it with the `--debug` flag. Other arguments can be passed normal. With the `--debug` flag specified, the `Gemini API` request and response payloads and other relevant data will be written to `stderr`. This output is in the form of JSON encoded structured logs. As the primary responses are written to `stdout` the debug component can easily be separated from the main content, for independent analysis, using standard `redirection` techniques.



