# magic8

Ask the Magic 8-Ball.

## Synopsis

```bash
tofu magic8 [question]
```

## Description

Ask the Magic 8-Ball for guidance on important architectural decisions. Includes classic responses plus tech-flavored extras.

## Arguments

| Argument | Description | Required |
|----------|-------------|----------|
| `question` | Your question (optional, just for fun) | No |

## Examples

Ask a question:

```bash
tofu magic8 "Should I refactor this code?"
# Output: LGTM.
```

Just shake the 8-ball:

```bash
tofu magic8
# Output: It depends.
```

## Responses

### Positive
- It is certain.
- It is decidedly so.
- Without a doubt.
- Yes, definitely.
- You may rely on it.
- As I see it, yes.
- Most likely.
- Outlook good.
- Yes.
- Signs point to yes.

### Neutral
- Reply hazy, try again.
- Ask again later.
- Better not tell you now.
- Cannot predict now.
- Concentrate and ask again.

### Negative
- Don't count on it.
- My reply is no.
- My sources say no.
- Outlook not so good.
- Very doubtful.

### Tech-Flavored
- Have you tried turning it off and on again?
- Works on my machine.
- That's a feature, not a bug.
- Ship it.
- LGTM.
- Needs more unit tests.
- Ask the senior dev.
- Check Stack Overflow.
- It depends.
- That's out of scope.
