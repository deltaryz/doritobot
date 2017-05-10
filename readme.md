# doritobot

Discord bot software written in Golang which implements several commands, most of them utterly useless.

Commands:
* `nazoupdate`: Reminds the channel that Nazo is adorable (the original command!)
* `echo`: Causes the bot to echo the given text, prepended with `echo:`
* `cb`: Cleverbot.IO API integration; forwards the given message to Cleverbot and posts his response
* `db`: Make a Derpibooru search query with the given tags and post a random image result.
* `h`: Responds with "h".
* `pvfmservers`: Lists the various PVFM streams with direct source links.

The `config.json` file contains the login data and various other properties you may wish to adjust. An example json is included in the repo.

Unless disabled, an HTTP endpoint is also provided on port 8080 directory `/chat`. To use it, provide a Discord channel ID and the desired message like so: `http://server.poop:8080/chat?id=6969696969&msg=memes`

And yes, Nazo is very much adorable.
