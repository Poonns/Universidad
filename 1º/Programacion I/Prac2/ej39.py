#Introducir un numero y mostras los primos menores a dicho numero
num = int(input('Introduce un numero: '))

for divisores in range (2, num):
    primo = True
    for i in range (2, divisores):
        if divisores % i == 0 and divisores % divisores == 0 and divisores % 1 == 0:
            primo = False
    
    if primo:
        print(divisores)