veces_gym = int(input('Introduce cuantas veces al año vas a ir al gimnasio: '))

#Tarjeta -------------------------------------------------------------------
tarjeta = 50

#Bonos---------------------------------------------------------------------
bonos = (veces_gym + 9) // 10 * 20

#Bonos y entradas----------------------------------------------------------
bonos_entradas = (veces_gym // 10) * 20 + (veces_gym % 10) * 3


if tarjeta <= bonos and tarjeta <= bonos_entradas:
    mejor_opcion = "tarjeta"
elif bonos <= tarjeta and bonos <= bonos_entradas:
    mejor_opcion = "bonos"
elif bonos_entradas <= tarjeta and bonos_entradas <= bonos:
    mejor_opcion = "bonos y entradas"


print(f'El precio de tarjetas sería {tarjeta} euros')
print(f'El precio de bonos sería {bonos} euros')
print(f'El precio de bonos y entradas sería {bonos_entradas} euros')
print(f'Recomendacion: {mejor_opcion}')
