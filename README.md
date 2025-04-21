# q

This is `q` (from `query` or `question`). It is a command-line `llm` interface built on Google's `Gemini` models. 

## Features

Using `q` provides the following features within your terminal:

* Non-blocking, yet conversational, prompting allowing natural, fluid usage within the terminal
  * There is no entering of a dedicated `repl` to define a session; leaving the terminal free to execute other commands between prompts while still maintaining the conversational context
* Session management enables easy stashing of, or switching to, the currently active, or a previously stashed session
  * Easily task switch without losing the current conversational context
* Fully scriptable and ideal for use in automation and `ci` pipelines
  * All configuration and session history is file or flag based, responses can be structured with a `schema` and activity indicators disabled
* Support for structured responses using custom `schemas`
  * Optionally, have responses returned using a custom `schema` as opposed to free-form text
* Personalisation of responses
  * Specify personal or contextual information and style preferences to tailor responses to specific circumstances
* Model customisation
  * Specify custom model configurations to fine-tune output

## Installation

To install `q`, download the appropriate tarball for your `os` from the [releases](https://github.com/comradequinn/q/releases/) page. Extract the binary and place it somewhere accessible to your `$PATH` variable. 

Optionally, you can use the below script to do that for you.

```bash
export VERSION="v1.0.0"; export OS="linux-amd64"; wget "https://github.com/comradequinn/q/releases/download/${VERSION}/q-${VERSION}-${OS}.tar.gz" && tar -xf "q-${VERSION}-${OS}.tar.gz" && rm -f "q-${VERSION}-${OS}.tar.gz" && chmod +x q && sudo cp q /usr/bin/
```

### Removal

To remove `q`, delete the binary from `/usr/bin` (or the location it was originally installed in)

## Usage

### Prompting and Initial Configuration

To chat with `q`, execute it with a prompt

```bash
q "how do I list all files in my current directory?"
```

> If this is the first time `q` has been executed, before answering it will prompt for an `Gemini API Key`, contextual information about the user and some default model preferences. `Gemini API Keys` are available free from [Google](https://aistudio.google.com/apikey). If you are unsure about what model preferences to use, use the suggested defaults as these are tuned for the expected use-case of `q`. The contextual information you provide about the user is stored on the host machine and included in API calls to `Gemini`. It is not stored or transmitted in any other form or for any other purpose, under any circumstances. You can update the `Gemini API Key`, contextual information about the user or the model preferences, at any time, by running `q --config` 

The result of the prompt will be displayed, as shown below.

```
To list all files in your current directory, you can use the following command in your terminal:

ls -a

This command will display all files, including hidden files (files starting with a dot).
```

To ask a follow up question, run `q` again with the required prompt.

```bash
q "I need timestamps in the output"
```

This will return something similar to the below, note how `q` understood the context of the question in relation to the previous prompt. 

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
q --new "what was my last question?"
```

This will return something similar to the below, indicating the loss of the previous context.

```
I have no memory of past conversations. Therefore, I don't know what your last question was.
```

### Session Management

A session is a thread of prompts and responses with the same context, effectively a conversation. A new session starts whenever `--new` (or `-n`) is passed along with the prompt to `q`. At this point, the previously active session is `stashed` and the passed prompt becomes the start of a new session.

To view your previously `stashed` sessions, run `q --list` (or `-l`). The sessions will be displayed in date order and include a snippet of the opening text of the prompt for ease of identification. The active session is also included in the output and prefixed with an asterix, in this case record `2`.

```bash
q --list
  #1 (April 15 2025): 'how do i list all files in my current directory?'
* #2 (April 15 2025): 'what was my last question?'
```

To restore a previous session, allowing you to continue that conversation as it was where you left off, run `q --restore #id` (or `-r`) where `#id` is the `#ID` in the `q --list` output. For example

```bash
q --restore 1
```

Running `q --list` again will now show the below; note how the asterix is now positioned at record `1`

```bash
q --list
* #1 (April 15 2025): 'how do i list all files in my current directory?'
  #2 (April 15 2025): 'what was my last question?'
```

Asking the prompt from earlier, of `q "what was my last question?"`, will now return the below, as that context has been restored.

```
Your last question was: "I need timestamps in the output".
```

To delete a single session, run `q --delete #id` (or `-d #id`) where `#id` is the `#ID` in the `q --list` output. To delete all sessions, run `q --delete-all`

### Structured Responses

By default, `q` will request responses structured as free-form text, which is a sensible format for conversational use. However, in many scenarios, it is preferable to have the output in a structured form. To this end, `q` allows you to specify a `json` based `schema` that will be used to form the response.

An example is shown below, note that `grounding` must be disabled to use a `schema`. This is a current stipulation of the `Gemini API`, not `q` itself.

```bash
q -n --grounding=false --schema='{"type":"object","properties":{"response":{"type":"string"}}}' "pick a colour of the rainbow"
```

This will return a response similar to the below.

```text
{
  "response": "Blue"
}
```

It may be preferable to store complex `schemas` in a file rather than declaring them inline. Standard command substitution techniques can be used to enable this. The example below shows how the same `schema` as defined inline above can instead be read from the file `./schema.json`.

```bash
q -n --grounding=false --schema="$(cat ./schema.json)" "pick a colour of the rainbow"
```

#### Scripting

A typical use of a `schema` is in scripting scenarios; enabling support for basic agentic tasks, general automation or for incorporation into `ci` pipelines. Having machine-readable output in a consistent form allows decisions to be safely taken on the output. 

The below extends the previous example to supress activity indicators, with the `--script` flag, and pipe the result into `jq`.

```bash
q -n --grounding=false --script --schema="$(cat ./schema.json)" "pick a colour of the rainbow" | jq
```

This will return a response similar to the below.

```json
{
  "response": "Blue"
}
```

## Model Configuration 

Using `q` you can set various model configuration options. These include `temperature`, `top-p` and token limits. These can be set in the configuration by running `q --config` and applied to all subsequent sessions. Alternatively, they can be set on a per-request basis by specifying them inline. An example of the latter is shown below.

```bash
q --temperature 0.1 --top-p 0.1 "how do I list all files in my current directory?"
```

The effect of the above will be to make the responses more determistic and favour correctness over 'imagination'. While the effects of `top-p` and `temperature` are out of the scope of this document, briefly and simplisticly; when the LLM is selecting the next token to include in its response, the value of `top-p` restricts the pool of potential next tokens that can be selected to the most probable subset. This derived by selecting the most probable one by one until the cumulative probabilty exceeds `p`. The  `temperature` value is then used to weight the probabilities in the resultings subset `p` to either level them out or emphasise their differences; making it less or more likely that the highest probability candidate will be chosen.

## Debugging

To inspect the underlying Gemini API traffic that is generated by `q`, run it with the `--debug` flag. Other arguments can be passed as normal however with the `--debug` flag specified the API payloads will also be printed to `stderr`. As the primary responses are written to `stdout` the debug component can easily be separated from the main content with standard redirection techniques.



