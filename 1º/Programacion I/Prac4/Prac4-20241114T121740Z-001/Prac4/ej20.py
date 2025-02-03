def sumar_fila(matriz):
    for pos_fila in range(len(matriz)):
        suma_elementos_filas = 0
        for pos_columna in range(len(matriz[0])):
            suma_elementos_filas += matriz[pos_fila][pos_columna] 
        print(f'Suma de la fila {pos_fila + 1}: {suma_elementos_filas}')
    return suma_elementos_filas
    
from modulomatrices import leerMatrizEnteros, mostrarMatriz

texto = input('Introduce el nombre de un fichero: ')
matriz = leerMatrizEnteros(texto)



print(sumar_fila(matriz))
