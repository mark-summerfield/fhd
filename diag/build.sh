lout -U -C/home/mark/fonts -D/home/mark/fonts -F/home/mark/fonts \
    -o db.ps main.lout && \
    embedfonts.py db.ps > /dev/null && \
    pspdf -W 324 -H 425 db.ps db.pdf > /dev/null && \
    pdf2svg db.pdf db.svg && rm -f db.ps db.pdf
