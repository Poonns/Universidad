filas_1 = int(input('Introduce el numero de filas de la matriz 1: '))
columnas_1 = int(input('Introduce el numero de columnas de la matriz 1: '))
matriz_1 = []

filas_2 = int(input('Introduce el numero de filas de la matriz 2: '))
columnas_2 = int(input('Introduce el numero de columnas de la matriz 2: '))
matriz_2 = []

matriz_resultante = []


for posicion_fila in range(filas_1):
    fila = []
    for elemento in range(columnas_1):
        fila.append(int(input(f'Introduce un elemento a la fila {posicion_fila}: ')))
    matriz_1.append(fila)

for fila in matriz_1:
    print(fila)


for posicion_fila in range(filas_2):
    fila = []
    for elemento in range(columnas_2):
        fila.append(int(input(f'Introduce un elemento a la fila {posicion_fila}: ')))
    matriz_2.append(fila)

for fila in matriz_2:
    print(fila)

if filas_1 == filas_2 and columnas_1 == columnas_2:
    for fila in range(len(matriz_1)):
        nueva_fila = []
        for columna in range(len(matriz_1[0])):
            nueva_fila.append(matriz_1[fila][columna] + matriz_2[fila][columna])
        matriz_resultante.append(nueva_fila)
else: 
    print('Las matrices no tienen las mismas dimensiones')

for fila in matriz_resultante:
    print(fila)