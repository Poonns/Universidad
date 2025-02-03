def sumar_diagonal(matriz):
    if len(matriz) != len(matriz[0]):
        return None
    
    suma_diagonal = 0
    for i in range(len(matriz)):
        suma_diagonal += matriz[i][i]
    return suma_diagonal

from modulomatrices import leerMatrizEnteros, mostrarMatriz

texto = input('Introduce el nombre de un fichero: ')
matriz = leerMatrizEnteros(texto)


if sumar_diagonal(matriz) == None:
    print('ERROR. La matriz no es cuadrada')
else:
    print(sumar_diagonal(matriz))