void permRec(unsigned i, vector<unsigned> &d, unsigned x){
    if(i == d.size()){
        int suma = 0;
        for(int t= 0; t < d.size(); t++){
            suma += d[t];
        }

        if(suma == x){
            for(int t = 0; t < d.size(), t++){
                cout << d[t] << " ";
            }
        }
        return;
    }
    
    for(int v = 1; v <= k; v++){
        d[i] = v;
        if(no_repe(d,i)){
            permRec(i+1, d, x);
        }
    }
    
}

void perm(unsigned x, unsigned n){
    vector<unsigned> v(n);
    permRec(0, v, x);
}