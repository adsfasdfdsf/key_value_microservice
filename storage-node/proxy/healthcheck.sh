#!/bin/sh
set -eu

update_upstream() {
    probe=/etc/nginx/conf.d/storage-upstream.next
    {
        echo 'upstream storage_backend {'
        echo '    zone storage_nodes 64k;'
        healthy=0
        for node in master replica1 replica2; do
            if wget -q -T 1 -O /dev/null "http://${node}:7070/health/ready" 2>/dev/null; then
                address=$(getent hosts "$node" | awk 'NR==1 {print $1}')
                if [ -n "$address" ]; then
                    echo "    server ${address}:7070 max_fails=1 fail_timeout=5s;"
                    healthy=1
                fi
            fi
        done
        if [ "$healthy" -eq 0 ]; then
            echo '    server 127.0.0.1:9;'
        fi
        echo '}'
    } > "$probe"
    current=/etc/nginx/conf.d/storage-upstream.conf
    if ! cmp -s "$probe" "$current"; then
        mv "$probe" "$current"
        if [ -s /var/run/nginx.pid ]; then
            nginx -t && nginx -s reload
        fi
    fi
    rm -f "$probe"
}

update_upstream
(
    while sleep 2; do
        update_upstream
    done
) &
