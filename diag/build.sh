lout -U -C/home/mark/fonts -D/home/mark/fonts -F/home/mark/fonts \
    -o db.ps main.lout && \
    embedfonts.py db.ps > /dev/null && \
    pspdf -W 288 -H 470 db.ps db.pdf > /dev/null && \
    pdf2svg db.pdf db.svg && rm -f db.ps db.pdf
