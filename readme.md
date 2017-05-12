# doritobot

Discord bot software written in Golang which implements several commands, most of them utterly useless.

[![Build Status](https://travis-ci.org/techniponi/doritobot.svg?branch=master)](https://travis-ci.org/techniponi/doritobot)

Commands:
* `nazoupdate`: Reminds the channel that Nazo is adorable (the original command!)
* `echo`: Causes the bot to echo the given text, prepended with `echo:` (this can be disabled!)
* `cb`: Cleverbot.IO API integration; forwards the given message to Cleverbot and posts his response
* `db`: Make a Derpibooru search query with the given tags and post a random image result.
* `h`: Responds with "h".
* `techgore`: Posts a random image from [/r/techsupportgore](http://reddit.com/r/techsupportgore)
* `snuggle`, `cuddle`, `hug`, `kiss`, `boop` (and others...): Interact with any member of the Snuggle Trinity (or their husbandos)!
* `gay`: Summon an adorable picture of certain ponies.

There are also several [PonyvilleFM](http://ponyvillefm.com) related commands, these are tested on doritobot before being ported to PVFM's [aura](https://github.com/PonyvilleFM/aura).
* `pvfmservers`: Lists the various PVFM streams with direct source links.

The `config.json` file contains the login data and various other properties you may wish to adjust. An example json is included in the repo.

Unless disabled, an HTTP endpoint is also provided on port 8080 directory `/chat`. To use it, provide a Discord channel ID and the desired message like so: `http://server.poop:8080/chat?id=6969696969&msg=memes`

And yes, Nazo is very much adorable.
