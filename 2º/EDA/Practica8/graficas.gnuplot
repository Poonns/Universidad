
set terminal png size 800,600 enhanced font "Arial,12"

set output 'costes.png'
set title "Comparación de Coste (Accesos al Grafo)"
set xlabel "Número de nodos (n)"
set ylabel "Accesos promedio"
set grid

set key top left box

plot 'voraz.dat' using 1:2 with linespoints pt 7 lw 2 title 'Voraz Medio', \
     'voraz.dat' using 1:3 with lines dt 2 lw 2 title 'Voraz Teórico (n^2)', \
     'bt.dat' using 1:2 with linespoints pt 5 lw 2 lc rgb "red" title 'Backtracking'

set output 'sol.png'
set title "Calidad de la Solución (Tiempo total del recorrido)"
set xlabel "Número de nodos (n)"
set ylabel "Tiempo total (segundos)"
set grid

set key top left box

plot 'voraz_sol.dat' using 1:2 with linespoints pt 7 lw 2 title 'Solución Voraz', \
     'bt_sol.dat' using 1:2 with linespoints pt 5 lw 2 lc rgb "red" title 'Solución Óptima (BT)'
