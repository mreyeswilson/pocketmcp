#!/bin/bash

TAGS=$(git tag)

for tag in $TAGS; do
    gh release delete $tag --yes || true
    git push origin --delete $tag || true
    git tag -d $tag

    echo "Versión $TAG eliminada"
done