num = int(input('Introduce un numero: '))
cont = 0
divisores = 2
while cont < num:
    primo = True
    for i in range (2, divisores):
        if divisores % i == 0 and divisores % divisores == 0 and divisores % 1 == 0:
            primo = False
    
    if primo: 
        cont += 1
        print(divisores)
    
    divisores += 1
    