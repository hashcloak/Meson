FROM %%KATZENPOST_AUTH%%
RUN apk update --no-cache && apk add --no-cache bash
ADD auth.entry.sh /entry.sh
ENTRYPOINT /entry.sh
