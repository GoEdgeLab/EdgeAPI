package models

import "encoding/json"

func (this *IPLibraryFile) DecodeCountries() []string {
	var countries = []string{}
	if IsNotNull(this.Countries) {
		err := json.Unmarshal(this.Countries, &countries)
		if err != nil {
			// ignore error
		}
	}
	return countries
}

func (this *IPLibraryFile) DecodeProvinces() [][2]string {
	var provinces = [][2]string{}
	if IsNotNull(this.Provinces) {
		err := json.Unmarshal(this.Provinces, &provinces)
		if err != nil {
			// ignore error
		}
	}
	return provinces
}

func (this *IPLibraryFile) DecodeCities() [][3]string {
	var cities = [][3]string{}
	if IsNotNull(this.Cities) {
		err := json.Unmarshal(this.Cities, &cities)
		if err != nil {
			// ignore error
		}
	}
	return cities
}

func (this *IPLibraryFile) DecodeTowns() [][4]string {
	var towns = [][4]string{}
	if IsNotNull(this.Towns) {
		err := json.Unmarshal(this.Towns, &towns)
		if err != nil {
			// ignore error
		}
	}
	return towns
}

func (this *IPLibraryFile) DecodeProviders() []string {
	var providers = []string{}
	if IsNotNull(this.Providers) {
		err := json.Unmarshal(this.Providers, &providers)
		if err != nil {
			// ignore error
		}
	}
	return providers
}

func (this *IPLibraryFile) DecodeEmptyValues() []string {
	var result = []string{}
	if IsNotNull(this.EmptyValues) {
		err := json.Unmarshal(this.EmptyValues, &result)
		if err != nil {
			// ignore error
		}
	}
	return result
}
