#!/bin/bash
URL="http://localhost:8080/public"
for i in {1..100}
do
   curl -s -o /dev/null -w "Solicitud #$i - Código: %{http_code}\n" $URL &
done
wait
echo "Prueba de carga finalizada."