# q

Named `q` (from `query` or `question`), `q` is a command-line `llm` interface built on Google's `Gemini 2.5` models. 

Using `q` greatly simplifies integrating LLMs into CI pipelines, scripts or other automation. For terminal users, `q` is also a simple, but powerful chat-based assistant.

## Features

Using `q` provides the following features:

* Interactive command-line chatbot
  * Non-blocking, yet conversational, prompting allowing natural, fluid usage within the terminal environment
    * The avoidance of a dedicated `repl` to define a session leaves the terminal free to execute other commands between prompts while still maintaining the conversational context
  * Session management enables easy stashing of, or switching to, the currently active, or a previously stashed session
    * Making it simple to quickly task switch without permanently losing the current conversational context
* Fully scriptable and ideal for use in automation and `ci` pipelines
  * All configuration and session history is file or flag based
  * API Keys are provided via environment variables
  * Support for structured responses using custom `schemas`
    * Basic schemas can be defined using a simple schema definition language
    * Complex schemas can be using OpenAPI Schema objects expressed as JSON and optionally defined in dedicated files for ease of management
  * Interactive-mode activity indicators can be disabled to aid effective redirection and piping
* Support for attaching files to prompts
  * Interrogate code, markdown and text files
* Personalisation of responses
  * Specify persistent, personal or contextual information and style preferences to tailor your responses
* Model configuration
  * Specify custom model configurations to fine-tune output

## Installation

To install `q`, download the appropriate tarball for your `os` from the [releases](https://github.com/comradequinn/q/releases/) page. Extract the binary and place it somewhere accessible to your `$PATH` variable. 

Optionally, you can use the below script to do that for you.

```bash
export VERSION="v1.2.0"; export OS="linux-amd64"; wget "https://github.com/comradequinn/q/releases/download/${VERSION}/q-${VERSION}-${OS}.tar.gz" && tar -xf "q-${VERSION}-${OS}.tar.gz" && rm -f "q-${VERSION}-${OS}.tar.gz" && chmod +x q && sudo mv q /usr/local/bin/
```

### API Keys

In order to use `q` you will require your `Gemini API Key`. If you do not already have one, these are available free from [Google](https://aistudio.google.com/apikey). 

Once you have the key, set and export it as the conventional environment variable for that value, `GEMINI_API_KEY`.

For convenience, you may wish to add this to your `~/.bashrc` file. An example is shown below

```bash
# file: ~/.bashrc

export GEMINI_API_KEY="myPriVatEApI_keY_1234567890"
```

### Removal

To remove `q`, delete the binary from `/usr/bin` (or the location it was originally to). You may also wish to delete its application directory that stores user preferences and chat history; this is located at `~/.q`.

## Usage

### Prompting

To chat with `q`, execute it with a prompt

```bash
q "how do I list all files in my current directory?"
```

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

### Personalisation

You can provide persistent, contextual information about yourself (or the running process) and preferred response styles, at any time, by running `q --config` and answering the prompts. Any information provided will then be implicitly included in all prompts sent to the `Gemini API` from that point on.

Any such information you provide is only stored on the host machine and included in calls to the `Gemini API`. It is not stored or transmitted in any other form or for any other purpose, under any circumstances. 

### Attaching Files

To attach a file to your prompt, use the `--file` (or `-f`) parameter passing the path to the file to include. An example is shown below.

```bash
q -n --file some-code.go "summarise this file"
```

### Grounding

Grounding is the term for verifying LLM responses with an external source, that source being `Google Search` in the case of `q`. By default this feature is enabled, but it can be disabled with `--grounding=false` as shown below.

```bash
q -n --grounding=false "how do I list all files in my current directory?"
```

### Scripting

When using the output of `q` in a script, it is advisable to supress activity indicators and other interactive output using the `--script` flag. This ensures a consistent output stream containing only response data.

The simple example below uses redirection to write the response to a file.

```bash
q -n --script "pick a colour of the rainbow" > colour.txt
```
This will result in a file similar to the below

```bash
# file: colour.txt
Blue
```

### Structured Responses

By default, `q` will request responses structured as free-form text, which is a sensible format for conversational use. However, in many scenarios, particuarly ci and scripting use-cases, it is preferable to have the output in a structured form. To this end, `q` allows you to specify a `schema` that will be used to format the response.

There are two methods of specifying a schema, either by using `QSF` (`q`'s `S`chema `F`ormat) or by providing a JSON based `OpenAPI schema object`. 

In either case, note that `grounding` must be disabled to use a `schema`. This is a current stipulation of the `Gemini API`, not `q` itself.

#### QSF (Q's Schema Format)

`QSF` provides a quick, simple and readable method of defining basic response schemas. It allows the definition of an arbitary number of `fields`, each with a `type` and an optional `description`. `QSF` can only be used to define non-hierarchical schemas, however this is often all that is needed for a substantial amount of structured response use-cases.

The most basic definition of a `QSF` schema, representing a single field response with no description is shown below

```bash
field-name:type
```
A more complex definition showing multiple fields, each with descriptions, is structured as follows.

```bash
field-name1:type1:description1,field-name2:type2:description2,...n
```

Providing a description can be useful for both the LLM and the user in understanding the purpose of the field. It can also reduce the amount of guidance needed in the main prompt itself to ensure response content is correctly assigned.

A simple example of execting `q` with a `QSF` defined schema is shown below.

```bash
q -n --grounding=false --script --schema='colour:string' "pick a colour of the rainbow"
```

This will return a response similar to the following.

```json
{
  "colour": "Blue"
}
```

#### Open API Schema

For more complex schemas, the definition can be provided as an [OpenAPI Schema Object](https://spec.openapis.org/oas/v3.0.3#schema-object-examples). A simple example is shown below.

```bash
q -n --grounding=false --script --schema='{"type":"object","properties":{"colour":{"type":"string", "description":"the selected colour"}}}' "pick a colour of the rainbow"
```

This will return a response similar to the following.

```json
{
  "colour": "Blue"
}
```

It may be preferable to store complex `schemas` in a file rather than declaring them inline. Standard command substitution techniques can be used to enable this. The example below shows how the same `schema` as defined inline above can instead be read from the file `./schema.json`.

```bash
q -n --grounding=false --schema="$(cat ./schema.json)" "pick a colour of the rainbow"
```

## Model Configuration 

Using `q` you can set various model configuration options. These include `temperature`, `top-p` and token limits. An example is shown below.

```bash
q --temperature 0.1 --top-p 0.1 "how do I list all files in my current directory?"
```

The effect of the above will be to make the responses more determistic and favour correctness over 'imagination'. 

While the effects of `top-p` and `temperature` are out of the scope of this document, briefly and simplisticly; when the LLM is selecting the next token to include in its response, the value of `top-p` restricts the pool of potential next tokens that can be selected to the most probable subset. This derived by selecting the most probable, one by one, until the cumulative probabilty exceeds the value of `p`. The `temperature` value is then used to weight the probabilities in that resulting subset to either level them out or emphasise their differences; making it less or more likely that the highest probability candidate will be chosen.

## Debugging

To inspect the underlying Gemini API traffic that is generated by `q`, run it with the `--debug` flag. Other arguments can be passed as normal however with the `--debug` flag specified the API payloads will also be printed to `stderr`. As the primary responses are written to `stdout` the debug component can easily be separated from the main content with standard redirection techniques.



