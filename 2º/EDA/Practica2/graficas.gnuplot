set terminal png
set xlabel "n"
set ylabel "Pasos / Tiempo"

##########################
# fa - Real + Teórico juntos
##########################
set title "Función fa - Caso medio"
set output 'fa.png'
plot "at_teo.dat" using 1:2 with lines title "fa teórico" lw 2 lc rgb "red", \
     "at_real.dat" using 1:2 with lines title "fa real" lw 2 lc rgb "blue"

##########################
# fb - Mejor caso (teórico + real)
##########################
set title "Función fb - Mejor caso"
set xlabel "n"
set ylabel "Tiempo/Coste"
set grid
set output 'fb_m.png'
plot "btm_teo.dat" using 1:2 with lines title "fb teórico mejor" lw 2 lc rgb "red", \
     "btm_real.dat" using 1:2 with lines title "fb real mejor" lw 2 lc rgb "blue"

##########################
# fb - Peor caso (teórico + real)
##########################
set title "Función fb - Peor caso"
set xlabel "n"
set ylabel "Tiempo/Coste"
set grid
set output 'fb_p.png'
plot "btp_teo.dat" using 1:2 with lines title "fb teórico peor" lw 2 lc rgb "red", \
     "btp_real.dat" using 1:2 with lines title "fb real peor" lw 2 lc rgb "blue"

##########################
# fb - Caso medio teórico
##########################
set title "Función fb - Caso medio Teórico"
set xlabel "n"
set ylabel "Tiempo/Coste"
set grid
set output 'fb_medio_teo.png'
plot "btmid_teo.dat" using 1:2 with lines title "fb teórico medio" lw 2 lc rgb "red"

##########################
# fb - Caso medio real
##########################
set title "Función fb - Caso medio Real"
set xlabel "n"
set ylabel "Tiempo/Coste"
set grid
set output 'fb_medio_real.png'
plot "btmid_real.dat" using 1:2 with lines title "fb real medio" lw 2 lc rgb "blue"
