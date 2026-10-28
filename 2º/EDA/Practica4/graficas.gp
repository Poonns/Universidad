set terminal pngcairo size 800,600 enhanced font 'Verdana,10'
set grid

set output 'merge_t.png'
set title "Mergesort: Coste en Pasos"
set xlabel "n"
set ylabel "Pasos"
set key left top

f(x) = a * x * (log(x)/log(2)) + b * x
fit f(x) 'merge_pasos.dat' via a, b

plot 'merge_pasos.dat' title "Empirico" with points pt 7 ps 1.2 lc rgb "blue", \
     f(x) title "Teorico (n log n)" with lines lw 2 lc rgb "red"

set output 'merge_r.png'
set title "Mergesort: Tiempo Real (-O2)"
set ylabel "Segundos"

g(x) = c * x * (log(x)/log(2)) + d * x
fit g(x) 'merge_t.dat' via c, d

plot 'merge_t.dat' title "Tiempo CPU" with linespoints pt 5 ps 1.0 lc rgb "green", \
     g(x) title "Ajuste n log n" with lines dt 2 lw 2 lc rgb "black"
