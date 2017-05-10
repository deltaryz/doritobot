# nazo-update

Discord bot software written in Golang which implements several commands, most of them utterly useless.

Commands:
* `nazoupdate`: Reminds the channel that Nazo is adorable (the original command!)
* `deltaspeak`: Causes the bot to echo the given text, prepended with `ds:`
* `cb`: Cleverbot.IO API integration; forwards the given message to Cleverbot and posts his response
* `db`: Make a Derpibooru search query with the given tags and post a random image result.

An HTTP endpoint is also provided on port 8080 directory `/chat`. To use it, provide a Discord channel ID and the desired message like so: `http://server.poop:8080/chat?id=6969696969&msg=memes`

And yes, Nazo is very much adorable.
