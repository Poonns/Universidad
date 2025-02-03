""" Disponemos de una matriz de dimensiones M × N (donde N es m´ultiplo de 12) que almacena el importe mensual
facturado por cada uno de los M comerciales de una empresa durante los ´ultimos N
12 a˜nos. Las columnas se
corresponden con los meses de forma consecutiva: la columna 0 es la de enero del primer a˜no, la columna 1 la de
febrero del primer a˜no, . . . , la columna 11 la de diciembre del primer a˜no, la columna 12 la de enero del segundo
a˜no y as´ı sucesivamente.

a) (1,5 puntos) Una funci´on, mejor_comercial_´ultimo_a~no, que reciba una matriz de facturaci´on como la
descrita anteriormente y devuelva el comercial (n´umero de 0 a M − 1) que ha facturado el mayor importe total
durante el ´ultimo a˜no, es decir, solamente durante los 12 ´ultimos meses. En caso de empate, puedes devolver
cualquiera de ellos.
Considerando la matriz de ejemplo, esta funci´on debe devolver 1. """

def mejor_comercial_ultimo_año(matriz):
    comercial_max_facturacion = -1
    indice_comercial = -1
    
    for i in range(len(matriz)):
        suma_facturacion = 0
        for j in range(12, len(matriz[0])):
            suma_facturacion += matriz[i][j]

        if suma_facturacion > comercial_max_facturacion:
            comercial_max_facturacion = suma_facturacion
            indice_comercial = i

    return indice_comercial

""" b) (1,5 puntos) Una funci´on, facturaci´on_total_por_mes, que reciba una matriz de facturaci´on y devuelva una
lista de 12 valores correspondientes al importe total facturado por mes por todos los comerciales y acumulando
todos los a˜nos.
Considerando la matriz de ejemplo, esta funci´on debe devolver la lista:
[10, 10, 14, 15, 17, 16, 17, 14, 12, 11, 10, 9] """

def facturacion_total_por_mes(matriz):
    lista = []

    for j in range(len(matriz[0]) // 2):
        suma = 0
        for i in range(len(matriz)):
            suma = matriz[i][j] + matriz[i][j+12]
            lista.append(suma)

    return lista