FROM postgres as base

WORKDIR /run

COPY run.sh run.sh

RUN chmod 700 run.sh

CMD [ "./run.sh" ]


FROM aws-cli:v2

WORKDIR /backups
COPY --from=base /backups .

ENTRYPOINT [ "aws", "s3", "sync" ]