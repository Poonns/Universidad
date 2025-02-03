def leer_alturas():
    lista = []
    altura = input('Introduce una altura: ')
    while altura != '':
        altura = int(altura)
        lista.append(altura)
        altura = input('Introduce otra altura: ')
    return lista

def mostrar_tramos_subida(lista):
    tramos = []
    tramo_actual = []
    for i in range(len(lista)):
        if len(tramo_actual) == 0:
            tramo_actual.append(lista[i])
        elif lista[i] > tramo_actual[-1]:
            tramo_actual.append(lista[i])
        else:
            if len(tramo_actual) > 1:
                tramos.append(tramo_actual)
            tramo_actual = []
    
    if len(tramo_actual) > 1:
        tramos.append(tramo_actual)
    
    return tramos

def mostrar_tramo(lista):
    numero_tramo = 1
    for tramo in lista:
        #Inicio y final del tramo
        principio_tramo = tramo[0]
        final_tramo = tramo[-1]

        #Longuitud del tramo
        loguitud_tramo = len(tramo)

        #Calcular el mayor desnivel
        mayor = tramo[0]
        menor = tramo[0]
        for altura in range(len(tramo)):
            if altura > mayor:
                mayor = tramo[altura]
            elif altura < menor:
                menor = tramo[altura]
        mayor_desnivel = mayor - menor

        
        

        
        

print('Ve introduciendo alturas sobre el nivel del mar, o una cadena vacia para acabar...')

lista = leer_alturas()
print(f'Lista de alturas: {lista}')

lista_tramo = mostrar_tramos_subida(lista)
print(f'Lista de tramos: {lista_tramo}')
print(len(lista_tramo))


if len(lista) < 2:
    print('Recorrido no válido con menos de dos puntos.')


