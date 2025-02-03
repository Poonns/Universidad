def leer_alturas():
    lista = []
    altura = input('Introduce una altura: ')
    while altura != '':
        altura = int(altura)
        lista.append(altura)
        altura = input('Introduce otra altura: ')
    return lista

def calcular_desniveles(lista):
    lista_desniveles = []
    for i in range(len(lista)-1):
        desnivel = lista[i+1] - lista[i] 
        lista_desniveles.append(desnivel)
    return lista_desniveles

def posiciom_maximo(lista):
    #CAMBIAR FUNCION! No es lo que pide el enunciado
    mayor_desnivel = 0
    for i in range(len(lista)):
        if mayor_desnivel <= lista[i]:
            mayor_desnivel = lista[i]
    return mayor_desnivel

print('Ve introduciendo alturas sobre el nivel del mar, o una cadena vacia para acabar...')
lista = leer_alturas()
print(lista)
lista_desniveles = calcular_desniveles(lista)
print(lista_desniveles)

if len(lista) < 2:
    print('Recorrido no válido con menos de dos puntos.')
elif len(lista_desniveles) < 0:
    print(f'Ningun kilometro es de subida')
else:
    mayor_desnivel = 0
    posicion_max = 0
    posicion_siguiente_max = 0
    for i in range(len(lista_desniveles)-1):
        if lista_desniveles[i] <= lista_desniveles[i+1]:
            mayor = lista_desniveles[i+1]
            posicion_max = i 
            posicion_siguiente_max = i+1
    
    print('Kilometro de mayor desnivel de subida:')
    print(f'\tEntre los puntos {posicion_max} y {posicion_siguiente_max}')
    print(f'\tDesnivel de {mayor}m')
 