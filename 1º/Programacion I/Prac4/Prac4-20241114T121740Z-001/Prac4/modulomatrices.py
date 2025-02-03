# -*- coding: iso-8859-15 -*-
#
# VERSION 1.0 (7-Noviembre-2012)
#
# AYUDA del módulo 'modulomatrices'. Define dos funciones, 
# que podemos utilizar de la siguiente forma:
#
# - matriz = leerMatrizEnteros(nombreFichero). Lee la matriz del fichero
#   indicado y la devuelve como una matriz de enteros. 
#
# - mostrarMatriz(matriz). Escribe en la pantalla los datos de la matriz.

def leerMatrizEnteros(nombreFichero):
    f=open(nombreFichero)
    im1=f.read()
    f.close()
    #convierte datos en matriz
    lin=im1.split('\n')
    mat=[]
    for l in lin:
        if l=='': break
        mat.append(list(map(int,l.split())))
    return mat

def mostrarMatriz(mat):
    for fila in mat:
        print("  ".join(list(map("{0:4d}".format,fila))))
