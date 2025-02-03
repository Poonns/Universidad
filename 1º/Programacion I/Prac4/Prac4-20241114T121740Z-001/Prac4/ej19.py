def sumar_elementos(matriz):
    suma_elementos = 0
    for fila in range(len(matriz)):
        for columna in range(len(matriz[0])):
            suma_elementos += matriz[fila][columna]
    return suma_elementos

from modulomatrices import leerMatrizEnteros, mostrarMatriz

texto = input('Introduce el nombre de un fichero: ')
matriz = leerMatrizEnteros(texto)



print(f'La suma de la matriz es: {sumar_elementos(matriz)}')