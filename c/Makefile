CC=gcc
CFLAGS=-Wall -Werror -Iinclude

# List your source files here
SRCDIR=src
SOURCES=$(wildcard $(SRCDIR)/*.c)
OBJECTS=$(SOURCES:.c=.o)

# Define executables for each source file
EXECUTABLES=$(SOURCES:$(SRCDIR)/%.c=%)

all: $(EXECUTABLES)

# Rule to build each executable
%: $(SRCDIR)/%.c
	$(CC) $(CFLAGS) -o $@ $<

clean:
	rm -f $(EXECUTABLES) $(OBJECTS)


test: $(EXECUTABLES)
	@./run_tests.sh
