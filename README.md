# Q

Q (from `Query` or `Question`) is a command-line LLM interface. It supports personalised responses and LLM configuration along with conversation management. It uses the Gemini 2.5 model by default.

## Installation

To install `q`, download the appropriate tarball for your OS from the [releases](https://github.com/comradequinn/q/releases/) page. Then extract the binary from it and place it some accessible to your `$PATH` variable. 

Optionally, you can use the below script to do that for you.

```bash
export VERSION="v1.0.3"; export OS="linux-amd64"; wget "https://github.com/comradequinn/q/releases/download/${VERSION}/q-${VERSION}-${OS}.tar.gz" && tar -xf "q-${VERSION}-${OS}.tar.gz" && rm -f "q-${VERSION}-${OS}.tar.gz" && chmod +x q && sudo cp q /usr/bin/
```

### Removal

To remove `q`, delete the binary from `/usr/bin` (or whatever location you installed it to)

## Usage

To use `q`, execute it with a prompt

```bash
q "how do I list all files in my current directory?"
```

If this is the first time `q` has been executed, before answering it will prompt for an Gemini API key and some personal information. Gemini API Keys are available free from [google](https://aistudio.google.com/apikey). 

The personal information you provide is stored on the host machine and included in API calls to Gemini. It is not stored or transmitted in any other form or for any other purpose, under any circumstances. 

You can update the API key or personal information at any time by running `q --config` 

Once the requested information has been provided, the result will be displayed, as shown below.

```
To list all files in your current directory, you can use the following command in your terminal:

ls -a

This command will display all files, including hidden files (files starting with a dot).
```

To ask a follow up question, run `q` again with the required prompt.

```bash
q "I need timestamps in the output"
```

This would return something similar to the below, note how `q` understood the context of the question in relation to the previous prompt. 

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

To delete a single session, run `q --delete #id` where `#id` is the `#ID` in the `q --list` output. To delete all sessions, run `q --delete-all`

## Configuration

Using `q` you can set various model configuration and debug options, to view these, and other available variables, run `q -h`
