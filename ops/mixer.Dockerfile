FROM %%KATZEN_SERVER%%
RUN apk update --no-cache && apk add --no-cache bash
ADD mixer.entry.sh /entry.sh
ENTRYPOINT /entry.sh
