# MILLION CHECKBOXES

## SUMMARY

This project is inspired by the One Million Checkboxes project as featured on Hacker News.

I wanted to take a different approach from that project, to better reflect more typical systems present in large 
enterprises.  Such as relational databases as the ultimate source of truth, and event queues for transmitting 
state changes.

This is an excellent project to demonstrate a variety of architecture styles to support high volume and
shared state.

This also gives an opportunity to explore different ways to handle conflict in
shared state.

## GOALS

There are three goals with this project:

1. Build a fun distributed checkbox app similar to the Million Checkboxes, but with more typical enterprise architecture.
2. Learn a new language to me, Go.
3. Experiment with and learn LLM coding assist tools, such as Claude Code and others.

## DISCUSSION

Make sure you review `CLAUDE.md` for the most up to date description of how to navigate and understand the project at a 
high level.

Note the various files under `/docs/`.  The architecture, goals, and scaling were original pre-coding design ideas, and 
the system has grown somewhat past some of those details.

Note also the `docs/claude/commands` folder, where I kept some of the written commands for the actual coding work I had 
claude do.

## NOTES ON USE OF LLMS

I would definitely not consider this "vibe coding". I am trying to build a production-ready system. I may not be there at 
any given time, but that is the goal.

My use of LLMs has been almost exlusively claude, but in a few different forms.

I use Opus 4 in the web to ask questions and learn about language features. This has been a truly excellent use of the 
tool, and has allowed me to learn key aspects of the language faster than I would have through google searches and 
blog readings. I did have to verify much of what was written, but that wasn't hard. The tool was excellent and reliable 
for this use.

I also used claude code to produce snippets for things like AWS setup scripts and best practices. In these cases I would 
copy the code over.

The rest of the use was Claude Code.  I used this to do a few bits of plumbing, like the logging, config, and error 
system. I also had it do some coding on the typical approach for dependency injection. This was just me making sure 
I had a database package that was driver-agnostic, so postgresql specific code or packages wouldnt leak out into the 
rest of the code.

I also used to do do a Proof Of Concept for a worker pool. Golang has a slightly different approach to this sort of 
"concurrency" than what I'm used to with Java or javascript. 

Another great use of the tool is the db docker launch script. Saved me half an hour of writing a good bash script 
to launch the docker image, wait for the db to come online, and then run the migrations. Great use to make that one 
time tool at `database/system/setup_database.sh`.
