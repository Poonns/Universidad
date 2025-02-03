def sumar_columna(matriz):
    for pos_columna in range(len(matriz[0])):
        suma_elementos_columna = 0
        for pos_fila in range(len(matriz)):
            suma_elementos_columna += matriz[pos_fila][pos_columna] 
        print(f'Suma de la columna {pos_columna + 1}: {suma_elementos_columna}')
    return suma_elementos_columna
    
from modulomatrices import leerMatrizEnteros, mostrarMatriz

texto = input('Introduce el nombre de un fichero: ')
matriz = leerMatrizEnteros(texto)



print(sumar_columna(matriz))